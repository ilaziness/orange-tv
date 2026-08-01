import { useState } from "react";
import { Link } from "react-router";
import { clientApi, errorMessage } from "@/lib/api";
import { useAuthStore } from "@/store/auth";
import { useLoginDialogStore } from "@/store/loginDialog";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Field, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Spinner } from "@/components/ui/spinner";
import { AlertCircleIcon } from "lucide-react";

export function LoginDialog() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const isOpen = useLoginDialogStore((s) => s.isOpen);
  const close = useLoginDialogStore((s) => s.close);
  const setToken = useAuthStore((s) => s.setToken);
  const loadProfile = useAuthStore((s) => s.loadProfile);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setSubmitting(true);
    try {
      const res = await clientApi.login(username, password);
      setToken(res.data.access_token);
      await loadProfile();
      close();
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Dialog
      open={isOpen}
      onOpenChange={(open) => {
        if (!open) close();
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>登录</DialogTitle>
          <DialogDescription>登录后享受更多功能</DialogDescription>
        </DialogHeader>
        {error ? (
          <Alert variant="destructive">
            <AlertCircleIcon />
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        ) : null}
        <form onSubmit={handleSubmit}>
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="login-dialog-username">用户名</FieldLabel>
              <Input
                id="login-dialog-username"
                placeholder="请输入用户名"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
              />
            </Field>
            <Field>
              <FieldLabel htmlFor="login-dialog-password">密码</FieldLabel>
              <Input
                id="login-dialog-password"
                type="password"
                placeholder="请输入密码"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </Field>
            <Button type="submit" disabled={submitting} className="w-full">
              {submitting ? <Spinner data-icon="inline-start" /> : null}
              登录
            </Button>
          </FieldGroup>
        </form>
        <p className="text-center text-sm text-muted-foreground">
          还没有账号？{" "}
          <Link
            to="/register"
            className="text-primary underline-offset-4 hover:underline"
            onClick={() => close()}
          >
            注册
          </Link>
        </p>
      </DialogContent>
    </Dialog>
  );
}
