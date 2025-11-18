function FeatureCard({ icon: Icon, title, description, ...rest }) {
  return (
    <article className="feature-card" {...rest} role="listitem">
      <div className="feature-icon" aria-hidden="true">
        <Icon size={24} strokeWidth={2} />
      </div>
      <div className="feature-body">
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
    </article>
  );
}

export default FeatureCard;
