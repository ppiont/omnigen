type FeatureCardProps = {
  icon: string
  title: string
  description: string
} & React.HTMLAttributes<HTMLElement>

function FeatureCard({ icon, title, description, ...rest }: FeatureCardProps) {
  return (
    <article className="feature-card" {...rest} role="listitem">
      <div className="feature-icon" aria-hidden="true">
        {icon}
      </div>
      <div className="feature-body">
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
    </article>
  );
}

export default FeatureCard;
