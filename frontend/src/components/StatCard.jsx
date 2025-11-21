function StatCard({ label, value, helper, trend, motionDelay = "0s", icon }) {
  const trendDirection = trend?.direction ?? "neutral";

  return (
    <article
      className="stat-card"
    >
      <div className="stat-card-head">
        <span className="stat-card-label">{label}</span>
        {icon && (
          <span className="stat-card-icon" aria-hidden="true">
            {icon}
          </span>
        )}
      </div>
      <p className="stat-card-value">{value}</p>
      {trend?.value && (
        <p
          className={`stat-card-trend ${
            trendDirection === "down" ? "is-down" : "is-up"
          }`}
        >
          {trend.value} <span>{trend.caption}</span>
        </p>
      )}
      {helper && <p className="stat-card-helper">{helper}</p>}
    </article>
  );
}

export default StatCard;
