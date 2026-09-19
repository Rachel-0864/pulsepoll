import { Link } from "react-router-dom";

export default function NotFound() {
  return <div className="empty-page"><h1>404</h1><p>That page doesn't exist.</p><Link className="button primary" to="/">Go home</Link></div>;
}