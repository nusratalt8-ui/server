package plan

func (s *Service) BroadcastPlanUpdate(userID string, plan int) {
	if s.panel != nil {
		s.panel.Emit("plan_updated", map[string]interface{}{
			"user_id": userID,
			"plan":    plan,
		})
	}
}
