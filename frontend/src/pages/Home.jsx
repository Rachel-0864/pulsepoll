import { Link } from "react-router-dom";
import { ArrowRight, Radio, ShieldCheck, Zap, BarChart3 } from "lucide-react";
import { useAuth } from "../context/AuthContext";

export default function Home() {
  const { user } = useAuth();

  return (
    <div className="home">
      <section className="hero">
        <div className="hero-badge"><span /> REAL-TIME POLLING</div>
        <h1>Ask. Vote.<br /><em>See it live.</em></h1>
        <p>Create a poll in seconds and watch responses appear in real time. No refresh. No waiting.</p>
        <div className="hero-actions">
          <Link className="button primary large" to={user ? "/create" : "/signup"}>Create a Poll <ArrowRight size={18} /></Link>
          <Link className="button secondary large" to="/poll/react-vs-vue">View Live Demo</Link>
        </div>
        <div className="hero-note"><Radio size={15} /> Live updates powered by WebSockets</div>
      </section>

      <section className="features">
        <Feature icon={<Zap />} title="Instant updates" text="Votes flow into the results without refreshing the page." />
        <Feature icon={<BarChart3 />} title="Clear results" text="See counts and percentages in an easy-to-read chart." />
        <Feature icon={<ShieldCheck />} title="Validated votes" text="The final backend validates every vote before processing it." />
      </section>
    </div>
  );
}

function Feature({ icon, title, text }) {
  return <div className="feature card"><div className="feature-icon">{icon}</div><h3>{title}</h3><p>{text}</p></div>;
}