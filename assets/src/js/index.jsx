import { createRoot } from "react-dom/client";
import App from "./App";
import { applyTheme } from "./themes";
import { ContextMenuProvider } from "./components/contextmenu/ContextMenu";

applyTheme(); // apply default until uiprefs loads
createRoot(document.getElementById("root")).render(
  <ContextMenuProvider>
    <App />
  </ContextMenuProvider>
);