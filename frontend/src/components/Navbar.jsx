import { Link, useNavigate } from "react-router-dom";
import { BarChart3, LogOut, Plus, LayoutDashboard } from "lucide-react";
import { useAuth } from "../context/AuthContext";

export default function Navbar() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate("/");
  };

  return (
    <header className="navbar">
      <Link className="brand" to="/">
        <span className="brand-mark"><BarChart3 size={20} /></span>
        <span>Pulse<span>Poll</span></span>
      </Link>
      <nav>
        {user ? (
          <>
            <Link to="/dashboard"><LayoutDashboard size={17} /> Dashboard</Link>
            <Link className="nav-primary" to="/create"><Plus size={17} /> Create Poll</Link>
            <button className="nav-user" onClick={handleLogout}>
              <span className="avatar">{user.name?.[0]?.toUpperCase() || "U"}</span>
              <span>{user.name}</span>
              <LogOut size={16} />
            </button>
          </>
        ) : (
          <>
            <Link to="/login">Login</Link>
            <Link className="nav-primary" to="/signup">Get Started</Link>
          </>
        )}
      </nav>
    </header>
  );
}