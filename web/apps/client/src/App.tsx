import { useEffect } from "react";
import { BrowserRouter } from "react-router";
import { AppRoutes } from "@/routes";
import { useSettingsStore } from "@/store/settings";
import { LoginDialog } from "@/components/auth/LoginDialog";

export default function App() {
  const loadSettings = useSettingsStore((s) => s.loadSettings);

  useEffect(() => {
    void loadSettings();
  }, [loadSettings]);

  return (
    <BrowserRouter>
      <AppRoutes />
      <LoginDialog />
    </BrowserRouter>
  );
}
