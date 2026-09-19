import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { UserPlus } from "lucide-react";
import { useAuth } from "../context/AuthContext";

export default function Signup() {
  const { signup } = useAuth();
  const navigate = useNavigate();

  const [form, setForm] = useState({
    name: "",
    email: "",
    password: "",
  });

  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const submit = async (e) => {
    e.preventDefault();

    setError("");
    setLoading(true);

    try {
      await signup(
        form.name,
        form.email,
        form.password
      );

      // Only navigate after signup succeeds
      navigate("/dashboard");
    } catch (err) {
      console.error("Signup error:", err);

      const message =
        err.response?.data?.error ||
        err.response?.data?.message ||
        err.message ||
        "Signup failed. Please try again.";

      setError(message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-page">
      <div className="auth-card card">
        <div className="auth-icon">
          <UserPlus />
        </div>

        <h1>Create your account</h1>
        <p>Start creating live polls.</p>

        {error && (
          <div className="alert">
            {error}
          </div>
        )}

        <form onSubmit={submit}>
          <label>Name</label>

          <input
            className="input"
            type="text"
            placeholder="Your name"
            value={form.name}
            onChange={(e) =>
              setForm({
                ...form,
                name: e.target.value,
              })
            }
            required
          />

          <label>Email</label>

          <input
            className="input"
            type="email"
            placeholder="you@example.com"
            value={form.email}
            onChange={(e) =>
              setForm({
                ...form,
                email: e.target.value,
              })
            }
            required
          />

          <label>Password</label>

          <input
            className="input"
            type="password"
            placeholder="At least 8 characters"
            minLength={8}
            value={form.password}
            onChange={(e) =>
              setForm({
                ...form,
                password: e.target.value,
              })
            }
            required
          />

          <button
            className="button primary full"
            type="submit"
            disabled={loading}
          >
            {loading ? "Creating Account..." : "Create Account"}
          </button>
        </form>

        <p className="auth-bottom">
          Already have an account?{" "}
          <Link to="/login">Login</Link>
        </p>
      </div>
    </div>
  );
}