"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { Check, LockKeyhole, UserRound } from "lucide-react";
import toast from "react-hot-toast";
import { authApi, coreApi, errorMessage } from "@/api/client";
import { useAuthStore } from "@/store/authStore";
import type { User } from "@/lib/types";
export function Profile() {
  const { user, setUser } = useAuthStore();
  const router = useRouter();
  const [saving, setSaving] = useState(false);
  const [changing, setChanging] = useState(false);
  const [passwordError, setPasswordError] = useState("");
  if (!user) return null;
  async function save(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = Object.fromEntries(new FormData(event.currentTarget));
    setSaving(true);
    try {
      const { data: result } = await coreApi.put<{ user: User }>("/me", data);
      setUser(result.user);
      toast.success("Your profile is up to date.");
    } catch (error) {
      toast.error(errorMessage(error));
    } finally {
      setSaving(false);
    }
  }
  async function changePassword(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    setPasswordError("");
    if (data.get("new_password") !== data.get("confirm")) {
      setPasswordError("Your new passwords don’t match.");
      return;
    }
    setChanging(true);
    try {
      await authApi.post("/change-password", Object.fromEntries(data));
      setUser(null);
      toast.success("Password updated. Sign in with your new password.");
      router.replace("/login");
    } catch (error) {
      setPasswordError(errorMessage(error));
    } finally {
      setChanging(false);
    }
  }
  return (
    <>
      <div className="page-heading">
        <div>
          <span className="eyebrow">YOUR CORNER OF LEARNYST</span>
          <h1>A little about you.</h1>
          <p>Keep your details up to date and your account secure.</p>
        </div>
      </div>
      <div className="profile-grid">
        <section className="panel">
          <div className="panel-heading">
            <UserRound size={19} />
            <h2>Personal details</h2>
          </div>
          <div className="profile-intro">
            <span className="avatar avatar-large">
              {user.first_name[0]}
              {user.last_name[0]}
            </span>
            <div>
              <strong>
                {user.first_name} {user.last_name}
              </strong>
              <p>{user.email}</p>
            </div>
          </div>
          <form onSubmit={save} className="stack-form">
            <div className="form-row">
              <label>
                First name
                <input
                  name="first_name"
                  defaultValue={user.first_name}
                  required
                  maxLength={100}
                />
              </label>
              <label>
                Last name
                <input
                  name="last_name"
                  defaultValue={user.last_name}
                  required
                  maxLength={100}
                />
              </label>
            </div>
            <label>
              Location <span className="optional">(optional)</span>
              <input
                name="location"
                defaultValue={user.location}
                placeholder="City, country"
                maxLength={255}
              />
            </label>
            <label>
              Email address
              <input value={user.email} disabled />
              <span className="field-help">
                Your email is used to sign in to your account.
              </span>
            </label>
            <button className="btn btn-primary" disabled={saving}>
              {saving ? "Saving…" : "Save changes"}
              <Check size={16} />
            </button>
          </form>
        </section>
        <section className="panel">
          <div className="panel-heading">
            <LockKeyhole size={19} />
            <h2>Account security</h2>
          </div>
          <p className="muted">
            Choose a unique password of at least 8 characters. Updating it will
            sign out all your sessions.
          </p>
          <form onSubmit={changePassword} className="stack-form">
            {passwordError && (
              <div className="notice error" role="alert">
                {passwordError}
              </div>
            )}
            <label>
              Current password
              <input
                name="old_password"
                type="password"
                autoComplete="current-password"
                required
                maxLength={72}
              />
            </label>
            <label>
              New password
              <input
                name="new_password"
                type="password"
                autoComplete="new-password"
                required
                minLength={8}
                maxLength={72}
              />
            </label>
            <label>
              Confirm new password
              <input
                name="confirm"
                type="password"
                autoComplete="new-password"
                required
                minLength={8}
                maxLength={72}
              />
            </label>
            <button className="btn btn-secondary" disabled={changing}>
              {changing ? "Updating…" : "Update password"}
              <LockKeyhole size={15} />
            </button>
          </form>
        </section>
      </div>
    </>
  );
}
