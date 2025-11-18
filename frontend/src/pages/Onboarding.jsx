import { Link } from "react-router-dom";
import {
  Activity,
  ArrowRightCircle,
  CheckCircle2,
  Compass,
  Layers,
  LayoutDashboard,
  Rocket,
  Sparkles,
  Workflow,
} from "lucide-react";
import "../styles/onboarding.css";

const whoAudience = [
  {
    title: "Pharmaceutical Marketing Teams",
    description:
      "Accelerate brand, product launch, and disease awareness campaigns with on-brand video content.",
    icon: <Rocket size={20} aria-hidden="true" />,
  },
  {
    title: "Medical Affairs & Education",
    description:
      "Produce training, MOA explainers, and conference-ready materials without lengthy production timelines.",
    icon: <Activity size={20} aria-hidden="true" />,
  },
  {
    title: "Agency & Partner Teams",
    description:
      "Deliver compliant pharmaceutical creative assets at scale for multiple brands and markets.",
    icon: <Layers size={20} aria-hidden="true" />,
  },
];

const workflowSteps = [
  {
    title: "Create",
    description:
      "Describe your campaign goals, target audience, regulatory guardrails, and upload existing brand assets.",
    icon: <Sparkles size={22} aria-hidden="true" />,
  },
  {
    title: "Generate",
    description:
      "PharmaGen assembles multi-scene videos with synchronized music, subtitles, and brand styling in minutes.",
    icon: <Workflow size={22} aria-hidden="true" />,
  },
  {
    title: "Refine",
    description:
      "Use the workspace timeline, script editor, and chat refinement tools to adjust messaging, pacing, or visuals.",
    icon: <Compass size={22} aria-hidden="true" />,
  },
  {
    title: "Export",
    description:
      "Download approved versions in 16:9, 9:16, and 1:1 formats, ready for paid media, organic, or medical education.",
    icon: <CheckCircle2 size={22} aria-hidden="true" />,
  },
];

const navigationGuide = [
  {
    title: "Dashboard",
    description:
      "Monitor active projects, access recently opened videos, and quickly jump back into your production pipeline.",
    icon: <LayoutDashboard size={20} aria-hidden="true" />,
  },
  {
    title: "Create",
    description:
      "Initiate a new campaign with creative parameters, brand directives, and advanced controls for pacing and tone.",
    icon: <Sparkles size={20} aria-hidden="true" />,
  },
  {
    title: "Library",
    description:
      "Browse generated outputs, track status, and reopen past campaigns. Filter by status or duration.",
    icon: <Layers size={20} aria-hidden="true" />,
  },
  {
    title: "Workspace",
    description:
      "Review a video’s timeline, adjust scripts, manage music, and finalize exports with professional controls.",
    icon: <Compass size={20} aria-hidden="true" />,
  },
  {
    title: "Settings",
    description:
      "Manage team profiles, access preferences, and integrate brand assets or compliance workflows.",
    icon: <Activity size={20} aria-hidden="true" />,
  },
];

const workspaceHighlights = [
  "Interactive timeline with dedicated tracks for video, music, and text overlays.",
  "Script editor with persistent versioning to refine messaging for HCP or patient audiences.",
  "Metadata panel summarizing duration, status, and distribution-ready assets.",
  "Seamless handoff between creative, regulatory, and marketing reviewers.",
];

function Onboarding() {
  return (
    <div className="onboarding-page">
      <header className="onboarding-hero">
        <div className="onboarding-hero-content">
          <div className="onboarding-hero-eyebrow">Welcome to PharmaGen</div>
          <h1>
            AI-Powered Video Creation for{" "}
            <span>Pharmaceutical Teams</span>
          </h1>
          <p>
            PharmaGen transforms complex pharmaceutical narratives into
            professional, multi-format video campaigns. This guide outlines how
            to navigate the platform and deliver compliant, engaging content
            faster.
          </p>
          <div className="onboarding-hero-actions">
            <Link className="onboarding-primary-cta" to="/create">
              Create Your First Video
              <ArrowRightCircle size={18} aria-hidden="true" />
            </Link>
            <Link className="onboarding-secondary-cta" to="/library">
              Explore Your Library
            </Link>
          </div>
        </div>
      </header>

      <section className="onboarding-block">
        <h2>Who PharmaGen Serves</h2>
        <p className="onboarding-lead">
          Purpose-built for regulated marketing and education teams who need
          agile storytelling with consistent brand execution.
        </p>
        <div className="onboarding-grid">
          {whoAudience.map((item) => (
            <article key={item.title} className="onboarding-card">
              <span className="onboarding-card-icon">{item.icon}</span>
              <div>
                <h3>{item.title}</h3>
                <p>{item.description}</p>
              </div>
            </article>
          ))}
        </div>
      </section>

      <section className="onboarding-block">
        <h2>How It Works</h2>
        <p className="onboarding-lead">
          A guided, compliant workflow takes you from brief to export-ready
          content in four streamlined steps.
        </p>
        <div className="onboarding-steps">
          {workflowSteps.map((step, index) => (
            <article key={step.title} className="onboarding-step">
              <div className="onboarding-step-index">
                {(index + 1).toString().padStart(2, "0")}
              </div>
              <div className="onboarding-step-icon">{step.icon}</div>
              <div className="onboarding-step-content">
                <h3>{step.title}</h3>
                <p>{step.description}</p>
              </div>
            </article>
          ))}
        </div>
      </section>

      <section className="onboarding-block">
        <h2>Platform Capabilities</h2>
        <p className="onboarding-lead">
          PharmaGen combines AI generation with pharma-grade review controls so
          teams can iterate quickly while maintaining compliance.
        </p>
        <ul className="onboarding-feature-list">
          <li>
            <Sparkles size={18} aria-hidden="true" />
            AI narrative planning tuned for pharmaceutical messaging and safety
            considerations.
          </li>
          <li>
            <Layers size={18} aria-hidden="true" />
            Simultaneous export in 16:9, 9:16, and 1:1 aspect ratios for omnichannel
            campaigns.
          </li>
          <li>
            <Workflow size={18} aria-hidden="true" />
            Batch generation for exploring variations by indication, market, or
            audience.
          </li>
          <li>
            <Compass size={18} aria-hidden="true" />
            Workspace collaboration for creative, regulatory, and brand teams to
            align in real time.
          </li>
          <li>
            <CheckCircle2 size={18} aria-hidden="true" />
            Compliance-ready structure with consistent captions, claims, and medical
            references.
          </li>
        </ul>
      </section>

      <section className="onboarding-block">
        <h2>Navigate the Platform</h2>
        <p className="onboarding-lead">
          Each area of PharmaGen focuses on a dedicated part of the production
          lifecycle.
        </p>
        <div className="onboarding-grid">
          {navigationGuide.map((item) => (
            <article key={item.title} className="onboarding-card">
              <span className="onboarding-card-icon">{item.icon}</span>
              <div>
                <h3>{item.title}</h3>
                <p>{item.description}</p>
              </div>
            </article>
          ))}
        </div>
      </section>

      <section className="onboarding-block">
        <h2>Inside the Workspace</h2>
        <p className="onboarding-lead">
          The workspace centralizes every control needed to finalize a video for
          launch or medical review.
        </p>
        <ul className="onboarding-highlight-list">
          {workspaceHighlights.map((highlight) => (
            <li key={highlight}>{highlight}</li>
          ))}
        </ul>
      </section>

      <section className="onboarding-block onboarding-cta-block">
        <h2>Ready to create?</h2>
        <p>
          Start with a prompt, refine with your team, and deliver regulatory-ready
          campaigns in a fraction of the traditional timeline.
        </p>
        <div className="onboarding-cta-actions">
          <Link className="onboarding-primary-cta" to="/create">
            Launch New Project
            <ArrowRightCircle size={18} aria-hidden="true" />
          </Link>
          <Link className="onboarding-secondary-cta" to="/dashboard">
            Return to Dashboard
          </Link>
        </div>
      </section>
    </div>
  );
}

export default Onboarding;

