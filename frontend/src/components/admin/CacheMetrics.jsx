import { useState, useEffect } from "react";
import { Activity, Database, DollarSign, Clock } from "lucide-react";
// Note: This component assumes a new API endpoint for cache stats which needs to be implemented in the backend
// For now, it will display placeholder data or can be connected once the endpoint is ready.

function MetricCard({ icon: Icon, label, value, subtext }) {
  return (
    <div className="metric-card">
      <div className="metric-icon">
        <Icon size={24} />
      </div>
      <div className="metric-content">
        <p className="metric-label">{label}</p>
        <p className="metric-value">{value}</p>
        {subtext && <p className="metric-subtext">{subtext}</p>}
      </div>
    </div>
  );
}

function CacheMetrics() {
  const [metrics, setMetrics] = useState({
    scriptHitRate: 0,
    sceneHitRate: 0,
    audioHitRate: 0,
    totalSavings: 0,
    cacheSize: 0,
    entriesCount: 0,
    avgResponseTime: 0
  });

  // In a real implementation, fetch metrics from backend API
  useEffect(() => {
    // Simulate fetching metrics
    const mockMetrics = {
      scriptHitRate: 32.5,
      sceneHitRate: 18.2,
      audioHitRate: 45.0,
      totalSavings: 1250.50,
      cacheSize: 2.4 * 1024 * 1024 * 1024, // 2.4 GB
      entriesCount: 15420,
      avgResponseTime: 45
    };
    setMetrics(mockMetrics);
  }, []);

  const formatBytes = (bytes) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  return (
    <div className="cache-metrics-dashboard">
      <h3>Cache Performance</h3>
      
      <div className="metrics-grid">
        <MetricCard 
          icon={Activity} 
          label="Hit Rates" 
          value={`${metrics.scriptHitRate}% (Scripts)`}
          subtext={`Scenes: ${metrics.sceneHitRate}% | Audio: ${metrics.audioHitRate}%`}
        />
        <MetricCard 
          icon={DollarSign} 
          label="Est. Savings" 
          value={`$${metrics.totalSavings.toFixed(2)}`}
          subtext="Based on avoided API calls"
        />
        <MetricCard 
          icon={Database} 
          label="Cache Size" 
          value={formatBytes(metrics.cacheSize)}
          subtext={`${metrics.entriesCount.toLocaleString()} entries`}
        />
        <MetricCard 
          icon={Clock} 
          label="Avg Response" 
          value={`${metrics.avgResponseTime}ms`}
          subtext="For cached content"
        />
      </div>
    </div>
  );
}

export default CacheMetrics;
