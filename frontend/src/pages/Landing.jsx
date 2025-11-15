import { Link } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext.jsx";
import FeatureCard from "../components/FeatureCard.jsx";
import "../styles/aurora.css";
import "../styles/landing.css";

const features = [
  {
    icon: "📱",
    title: "Multi-Format Export",
    description:
      "Generate videos in 16:9, 9:16, and 1:1 aspect ratios simultaneously for all platforms.",
  },
  {
    icon: "🎨",
    title: "Brand Consistency",
    description:
      "Apply your brand colors, fonts, and style guidelines across every variation automatically.",
  },
  {
    icon: "⚡",
    title: "A/B Test Ready",
    description:
      "Create multiple creative variations instantly to see which concepts drive the best performance.",
  },
  {
    icon: "💰",
    title: "Cost Efficient",
    description:
      "Generate professional-quality videos at under $2 per minute with optimized AI pipelines.",
  },
];

const steps = [
  {
    id: "01",
    title: "Brief",
    description:
      "Describe your product, upload assets, and set brand guidelines plus creative direction.",
  },
  {
    id: "02",
    title: "Generate",
    description:
      "Our pipeline creates ad variations with synced audio, text overlays, and on-brand styling.",
  },
  {
    id: "03",
    title: "Export",
    description:
      "Download production-ready videos in every format, ready for any advertising platform.",
  },
];

const stats = [
  {
    label: "Videos Generated",
    value: "10K+",
    helper: "For leading brands",
  },
  {
    label: "Success Rate",
    value: "98%",
    helper: "Across all renders",
  },
  {
    label: "Multi-Format",
    value: "16:9 · 9:16 · 1:1",
    helper: "Outputs in one pass",
  },
];

function Landing() {
  const { isAuthenticated } = useAuth();

  return (
    <main className="landing-page">
      <section className="landing-section hero-shell">
        <div className="container">
          <div className="hero-section aurora-field">
            <div
              className="hero-content"
              data-motion="rise"
              style={{ "--motion-delay": "0.05s" }}
            >
              <p className="hero-kicker">Ad Creative Pipeline</p>
              <h1 className="hero-title">Create Video Ads at Scale.</h1>
              <p className="hero-subtitle">
                Generate professional product videos and ad creatives in
                minutes. Multi-format output, brand-consistent styling, and
                production ready quality every time.
              </p>
              <div className="hero-ctas">
                {isAuthenticated ? (
                  <>
                    <Link className="btn btn-primary" to="/create">
                      Try Now
                    </Link>
                    <Link className="btn btn-secondary" to="/dashboard">
                      Dashboard
                    </Link>
                  </>
                ) : (
                  <>
                    <Link className="btn btn-primary" to="/signup">
                      Start Creating
                    </Link>
                    <Link className="btn btn-secondary" to="/login">
                      Login
                    </Link>
                  </>
                )}
              </div>
            </div>
          </div>
        </div>
      </section>

      <section
        className="landing-section feature-strip"
        aria-label="Feature list"
      >
        <div className="container">
          <div
            className="section-heading"
            data-motion="rise"
            style={{ "--motion-delay": "0.05s" }}
          >
            <p className="hero-kicker">Capabilities</p>
            <h2>Purpose-built for video ads.</h2>
            <p>
              Scale your ad production without scaling headcount. Omnigen keeps
              every format, brand rule, and storyboard move in sync.
            </p>
          </div>
          <div className="feature-grid" role="list">
            {features.map((feature, index) => (
              <FeatureCard
                key={feature.title}
                {...feature}
                data-motion="rise"
                style={{ "--motion-delay": `${0.1 * (index + 1)}s` }}
              />
            ))}
          </div>
        </div>
      </section>

      <section
        className="landing-section steps-section"
        aria-label="Workflow steps"
      >
        <div className="container">
          <div
            className="section-heading"
            data-motion="rise"
            style={{ "--motion-delay": "0.05s" }}
          >
            <p className="hero-kicker">Flow</p>
            <h2>Ship ads in three deliberate moves.</h2>
            <p>
              Each phase keeps your creative direction, compliance needs, and
              output formats aligned.
            </p>
          </div>
          <div className="steps-grid">
            {steps.map((step, index) => (
              <article
                key={step.id}
                className="step-card"
                data-motion="rise"
                style={{ "--motion-delay": `${0.1 * (index + 1)}s` }}
              >
                <span className="step-number">{step.id}</span>
                <h3>{step.title}</h3>
                <p>{step.description}</p>
              </article>
            ))}
          </div>
        </div>
      </section>

      <section
        className="landing-section stats-section"
        aria-label="Platform stats"
      >
        <div className="container">
          <div className="stats-grid">
            {stats.map((stat, index) => (
              <article
                key={stat.label}
                className="stat-card"
                data-motion="rise"
                style={{ "--motion-delay": `${0.1 * (index + 1)}s` }}
              >
                <p className="stat-label">{stat.label}</p>
                <p className="stat-value">{stat.value}</p>
                <p className="stat-helper">{stat.helper}</p>
              </article>
            ))}
          </div>
        </div>
      </section>

      <section className="landing-section">
        <div className="container">
          <div className="final-cta">
            <div className="final-cta-inner">
              <p className="hero-kicker">Takeoff</p>
              <h2>Ready to scale your ad production?</h2>
              <p>
                Join teams using Omnigen to create hundreds of ad variations in
                the time it used to take to produce one.
              </p>
              <Link className="btn btn-primary" to="/signup">
                Get Started Free
              </Link>
            </div>
          </div>
        </div>
      </section>
    </main>
  );
}

export default Landing;
