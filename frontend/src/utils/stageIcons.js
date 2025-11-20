import {
  FileText,
  Wand2,
  Film,
  Music,
  Scissors,
  CheckCircle,
  XCircle,
  Sparkles
} from 'lucide-react';

/**
 * Maps job stage names to lucide icons and colors
 * @param {string} stage - The current stage name (e.g., 'script_generating', 'scene_2_generating')
 * @returns {Object} Icon component and color
 */
export function getStageIcon(stage) {
  // Handle scene-specific stages (e.g., 'scene_1_generating', 'scene_2_complete')
  if (stage.startsWith('scene_') && stage.includes('_generating')) {
    return {
      icon: Wand2,
      color: '#8b5cf6', // Purple
      emoji: '🪄'
    };
  }

  if (stage.startsWith('scene_') && stage.includes('_complete')) {
    return {
      icon: Film,
      color: '#10b981', // Green
      emoji: '🎬'
    };
  }

  // Stage-specific mappings
  const stageMap = {
    'script_generating': {
      icon: FileText,
      color: '#3b82f6', // Blue
      emoji: '📝'
    },
    'script_complete': {
      icon: FileText,
      color: '#10b981', // Green
      emoji: '✅'
    },
    'audio_generating': {
      icon: Music,
      color: '#ec4899', // Pink
      emoji: '🎵'
    },
    'audio_complete': {
      icon: Music,
      color: '#10b981', // Green
      emoji: '🎶'
    },
    'composing': {
      icon: Scissors,
      color: '#f59e0b', // Amber
      emoji: '✂️'
    },
    'complete': {
      icon: CheckCircle,
      color: '#10b981', // Green
      emoji: '✅'
    },
    'failed': {
      icon: XCircle,
      color: '#ef4444', // Red
      emoji: '❌'
    }
  };

  // Return matched stage or default
  return stageMap[stage] || {
    icon: Sparkles,
    color: '#6366f1', // Indigo
    emoji: '✨'
  };
}

/**
 * Determines if a stage should show the dancing animation
 * @param {string} stage - The current stage name
 * @returns {boolean} True if stage is an active generating stage
 */
export function shouldDance(stage) {
  return (
    stage.includes('_generating') ||
    stage === 'composing'
  );
}

/**
 * Gets a user-friendly display name for a stage
 * (Fallback for when backend doesn't provide display_name)
 * @param {string} stage - The stage name
 * @returns {string} Display name
 */
export function getStageDisplayName(stage) {
  // Handle scene stages
  if (stage.startsWith('scene_')) {
    const match = stage.match(/scene_(\d+)_(generating|complete)/);
    if (match) {
      const sceneNum = match[1];
      const action = match[2] === 'generating' ? 'Generating' : 'Complete';
      return `${action} Scene ${sceneNum}`;
    }
  }

  const displayNames = {
    'script_generating': 'Generating Script',
    'script_complete': 'Script Ready',
    'audio_generating': 'Generating Audio',
    'audio_complete': 'Audio Ready',
    'composing': 'Composing Final Video',
    'complete': 'Complete',
    'failed': 'Failed'
  };

  return displayNames[stage] || stage.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
}
