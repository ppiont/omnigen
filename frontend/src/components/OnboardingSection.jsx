import { BookOpen, Compass, LayoutDashboard, Sparkles } from "lucide-react";
import PropTypes from "prop-types";
import { Link } from "react-router-dom";

function OnboardingSection({ className = "" }) {
  const points = [
    {
      icon: <Sparkles size={18} />,
      title: "Purpose-built for Pharma",
      description:
        "Create compliant-ready marketing, educational, and product launch videos tailored for pharmaceutical brands.",
    },
    {
      icon: <LayoutDashboard size={18} />,
      title: "AI-Powered Workflow",
      description:
        "Describe your campaign once and let PharmaGen generate, refine, and export professional-grade video variations.",
    },
    {
      icon: <Compass size={18} />,
      title: "Guided Workspace",
      description:
        "Edit scenes, manage scripts, review timelines, and export deliverables from a single workspace.",
    },
  ];

  return (
    <section className={`onboarding-section ${className}`.trim()}>
      <div className="onboarding-header">
        <BookOpen size={24} aria-hidden="true" />
        <div>
          <h2 className="onboarding-title">Getting Started with PharmaGen</h2>
          <p className="onboarding-subtitle">
            PharmaGen helps pharmaceutical teams produce highly polished video
            content without a production crew. Follow this guided workflow to
            get value in minutes.
          </p>
        </div>
      </div>

      <div className="onboarding-points">
        {points.map((point) => (
          <article key={point.title} className="onboarding-point">
            <span className="onboarding-point-icon" aria-hidden="true">
              {point.icon}
            </span>
            <div className="onboarding-point-content">
              <h3>{point.title}</h3>
              <p>{point.description}</p>
            </div>
          </article>
        ))}
      </div>

      <div className="onboarding-actions">
        <Link to="/onboarding" className="onboarding-learn-more">
          Learn More
        </Link>
      </div>
    </section>
  );
}

OnboardingSection.propTypes = {
  className: PropTypes.string,
};

export default OnboardingSection;

