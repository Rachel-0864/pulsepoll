import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from "recharts";

export default function PollResultsChart({ options }) {
  const data = options.map(o => ({ name: o.text, votes: o.count }));
  const total = options.reduce((sum, o) => sum + o.count, 0);

  return (
    <div className="results-wrap">
      <div className="results-total">
        <div><strong>{total}</strong><span>Total votes</span></div>
        <span className="live-pill">● LIVE</span>
      </div>
      <div className="chart">
        <ResponsiveContainer width="100%" height={300}>
          <BarChart data={data} layout="vertical" margin={{ top: 10, right: 20, left: 20, bottom: 10 }}>
            <CartesianGrid strokeDasharray="3 3" horizontal={false} />
            <XAxis type="number" allowDecimals={false} />
            <YAxis type="category" dataKey="name" width={110} tick={{ fontSize: 13 }} />
            <Tooltip cursor={{ fill: "rgba(99,102,241,.06)" }} />
            <Bar dataKey="votes" radius={[0, 7, 7, 0]} fill="#6366f1" />
          </BarChart>
        </ResponsiveContainer>
      </div>
      <div className="percentage-list">
        {options.map(o => {
          const pct = total ? Math.round((o.count / total) * 100) : 0;
          return (
            <div className="percentage-row" key={o.id}>
              <div><span>{o.text}</span><strong>{pct}%</strong></div>
              <div className="progress"><span style={{ width: `${pct}%` }} /></div>
            </div>
          );
        })}
      </div>
    </div>
  );
}