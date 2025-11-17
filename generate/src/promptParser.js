import OpenAI from 'openai';

/**
 * OpenAI-based prompt parser and content planner
 */
export class PromptParser {
  constructor(apiKey) {
    this.client = new OpenAI({ apiKey });
  }

  /**
   * Parse a user prompt and plan video content
   * @param {string} userPrompt - The user's video generation prompt
   * @param {number} numClips - Number of clips to generate (1 for single, >1 for sequence)
   * @param {Object} creativeOptions - Style, tone, tempo settings
   * @param {string} modelGuidance - Optional model-specific prompt guidance
   * @returns {Promise<Array>} Array of scene objects with prompts and metadata
   */
  async planContent(userPrompt, numClips = 1, creativeOptions = {}, modelGuidance = null) {
    console.log(`\n🤖 Planning content with OpenAI GPT-4o...`);
    console.log(`   Requested clips: ${numClips}`);
    if (creativeOptions.style || creativeOptions.tone || creativeOptions.tempo) {
      console.log(`   Creative direction: ${creativeOptions.style || 'default'} style, ${creativeOptions.tone || 'default'} tone, ${creativeOptions.tempo || 'medium'} tempo`);
    }
    if (modelGuidance) {
      console.log(`   📚 Using model-specific prompt optimization`);
    }

    const systemPrompt = numClips === 1
      ? this.getSingleClipSystemPrompt(creativeOptions, modelGuidance)
      : this.getMultiClipSystemPrompt(numClips, creativeOptions, modelGuidance);

    console.log(`   Using ${numClips === 1 ? 'SINGLE' : 'MULTI'} clip prompt`);

    try {
      // Higher temperature for more creative interpretations
      // 0.7 = balanced, 0.9 = more creative, 1.0 = very creative
      const temperature = creativeOptions.creative ? 0.9 : 0.7;

      const response = await this.client.chat.completions.create({
        model: 'gpt-4o',
        messages: [
          { role: 'system', content: systemPrompt },
          { role: 'user', content: userPrompt }
        ],
        response_format: { type: 'json_object' },
        temperature: temperature,
      });

      const result = JSON.parse(response.choices[0].message.content);

      console.log(`✅ Planned ${result.scenes.length} scene(s)\n`);

      // Validate: Ensure we got the right number of scenes
      if (result.scenes.length !== numClips) {
        console.log(`⚠️  Warning: AI returned ${result.scenes.length} scenes but ${numClips} were requested`);

        if (numClips === 1 && result.scenes.length > 1) {
          // For single clip requests, combine all scenes into one
          console.log(`   Combining all scenes into a single clip...`);
          const combinedPrompt = result.scenes.map(s => s.prompt).join(' ');
          const combinedDesc = result.scenes.map(s => s.description).join(', then ');
          return [{
            prompt: combinedPrompt,
            description: combinedDesc,
            duration: '8-10 seconds'
          }];
        } else if (result.scenes.length > numClips) {
          // Too many scenes - truncate
          console.log(`   Truncating to ${numClips} scene(s)...`);
          return result.scenes.slice(0, numClips);
        } else {
          // Too few scenes - pad with duplicates
          console.log(`   Padding to ${numClips} scene(s)...`);
          while (result.scenes.length < numClips) {
            result.scenes.push({...result.scenes[result.scenes.length - 1]});
          }
        }
      }

      return result.scenes;
    } catch (error) {
      console.error('Error calling OpenAI API:', error.message);
      throw error;
    }
  }

  getSingleClipSystemPrompt(creativeOptions = {}, modelGuidance = null) {
    const styleGuide = this.getStyleGuide(creativeOptions);
    const modelSection = modelGuidance ? `\nVIDEO MODEL SPECIFIC GUIDANCE:\n${modelGuidance}\n\nIMPORTANT: Optimize your prompts according to the model guidance above. Incorporate the recommended techniques and structures.\n` : '';
    const marketingFramework = this.getMarketingFramework(creativeOptions);
    const cinematographyGuide = this.getCinematographyGuide(creativeOptions);
    const platformOptimization = this.getPlatformOptimization(creativeOptions);

    return `You are an expert video director and cinematographer specializing in high-converting video advertisements. Convert the user's prompt into a compelling, cinematic 8-10 second video clip optimized for advertisement use.

${marketingFramework}
${styleGuide}
${cinematographyGuide}
${platformOptimization}
${modelSection}
CRITICAL RULES:
- Return EXACTLY ONE scene only
- Create commercial-quality, polished, attention-grabbing content
- Design for advertisement potential (can be used standalone or as part of a sequence)
- Include detailed cinematic direction
- FOLLOW THE STYLE GUIDE ABOVE for all creative decisions

Your task:
1. Identify the core concept from user's prompt
2. Create a compelling, visually rich video prompt (max 150 words)
3. Break down the shot into structured components (camera, lighting, subject, mood, pacing)
4. Make it attention-grabbing and professional

Return JSON with structured scene data:
{
  "scenes": [
    {
      "prompt": "Complete cinematic prompt - detailed, visual, commercial-quality",
      "description": "Brief high-level description",
      "duration": "8-10 seconds",
      "structure": {
        "camera": "Camera movement and angles (e.g., 'Smooth push-in from wide to close-up, handheld follow shot')",
        "lighting": "Lighting setup and mood (e.g., 'Golden hour natural light, dramatic shadows, warm tones')",
        "subject": "Main subject and action (e.g., 'Product rotating on pedestal, person using device')",
        "mood": "Emotional tone (e.g., 'Inspiring, energetic, calm, luxurious')",
        "pacing": "Timing and rhythm (e.g., 'Slow motion open, quick cuts midway, hold on final frame')",
        "style": "Visual aesthetic (e.g., 'Cinematic color grade, shallow depth of field, modern minimalist')"
      }
    }
  ]
}

The "scenes" array MUST contain exactly 1 element with complete structure data.`;
  }

  getMultiClipSystemPrompt(numClips, creativeOptions = {}, modelGuidance = null) {
    const styleGuide = this.getStyleGuide(creativeOptions);
    const modelSection = modelGuidance ? `\nVIDEO MODEL SPECIFIC GUIDANCE:\n${modelGuidance}\n\nIMPORTANT: Optimize your prompts according to the model guidance above. Incorporate the recommended techniques and structures.\n` : '';
    const marketingFramework = this.getMarketingFramework(creativeOptions);
    const cinematographyGuide = this.getCinematographyGuide(creativeOptions);
    const platformOptimization = this.getPlatformOptimization(creativeOptions);

    return `You are an expert video director and cinematographer specializing in high-converting video advertisements. Convert the user's prompt into a cohesive ${numClips}-part video sequence with narrative continuity and AIDA marketing flow.

${marketingFramework}
${styleGuide}
${cinematographyGuide}
${platformOptimization}
${modelSection}
CRITICAL RULES:
- Return EXACTLY ${numClips} scenes (no more, no less)
- Create a flowing STORY across all clips - each builds on the previous
- Maintain visual and narrative continuity (consistent style, lighting transitions, story progression)
- Each clip should be 8-10 seconds and work as part of the whole sequence
- Design for advertisement potential but prioritize storytelling

PROGRESSION REQUIREMENTS (CRITICAL - AVOID "SAME CLIP" SYNDROME):
Each clip MUST differ in at least 3 of these dimensions:
1. CAMERA POSITION: Change angle, distance, or movement (e.g., wide → medium → close-up)
2. ACTION PROGRESSION: Subject must DO different things (e.g., walking → running → celebrating)
3. ENVIRONMENT: Change setting, background, or location (e.g., indoor → outdoor → cityscape)
4. LIGHTING/TIME: Shift lighting mood or time of day (e.g., dawn → noon → dusk)
5. EMOTIONAL ARC: Character/viewer emotion should evolve (e.g., curious → excited → satisfied)

Story Structure Framework for ${numClips} clips:
${numClips === 2 ? `
- Clip 1: ESTABLISH (8-10s) - Set the scene, introduce concept/character, establish mood
- Clip 2: DEVELOP & RESOLVE (8-10s) - Progress the story, reveal outcome, memorable ending` :
numClips === 3 ? `
- Clip 1: BEGINNING (8-10s) - Establish setting, character, or concept. Set the tone
  → Example: Wide shot, subject introduced in static environment, calm lighting
- Clip 2: MIDDLE (8-10s) - Develop the narrative, show transformation or journey
  → Example: Medium shot, subject in ACTION (moving/interacting), environment changes, lighting shifts
- Clip 3: END (8-10s) - Resolution, payoff, emotional or visual climax
  → Example: Close-up or dramatic angle, subject reaches goal/transformation, lighting dramatic/resolved

ANTI-PATTERN: Do NOT create 3 clips showing the same scene from slightly different angles.
CORRECT PATTERN: Create 3 distinct moments in a progressing story (setup → action → payoff).` :
numClips === 4 ? `
- Clip 1: SETUP (8-10s) - Introduce world, character, or concept
- Clip 2: DEVELOPMENT (8-10s) - Build on setup, add complexity or movement
- Clip 3: CLIMAX (8-10s) - Peak moment, transformation, or key reveal
- Clip 4: RESOLUTION (8-10s) - Satisfying conclusion, emotional payoff` :
`
- First clip: ESTABLISH - Set scene and mood
- Middle clips: BUILD - Progress the story beat by beat
- Final clip: RESOLVE - Satisfying conclusion`}

Continuity Guidelines:
- Visual consistency: Keep similar color palette, lighting style across clips
- Smooth transitions: End each clip in a way that flows into the next
- Story progression: Each clip advances the narrative logically
- Maintain mood: Emotional tone should evolve naturally across sequence

EXAMPLE (Kobe Bryant + Cocoa Puffs, 3 clips):
❌ BAD (3 variations of same moment):
  Clip 1: "Kobe in kitchen holding Cocoa Puffs, warm lighting, medium shot"
  Clip 2: "Kobe in kitchen with Cocoa Puffs, warm lighting, close-up"
  Clip 3: "Kobe in kitchen pouring Cocoa Puffs, warm lighting, wide shot"
  → Problem: Same location, same action, only camera changes

✅ GOOD (Story progression):
  Clip 1: "Kobe in locker room post-game, exhausted. Close-up of hands unwrapping Cocoa Puffs box. Dim fluorescent lighting, medium shot"
  Clip 2: "Kobe in modern kitchen at home, energized, pouring Cocoa Puffs into bowl with confident smile. Morning golden hour light streaming through windows, wide shot"
  Clip 3: "Kobe courtside with Cocoa Puffs bowl, eating and laughing with teammates. Dynamic action, bright arena lights, handheld close-up on his satisfied expression, box visible in background"
  → Progression: Tired → Energized → Champion, Location changes, Action evolves, Lighting shifts

Your task:
1. Identify the core concept/story from user's prompt
2. Extract VISUAL CONSTANTS that should remain consistent (e.g., color palette, main subject identity, overall aesthetic)
3. Break it into ${numClips} connected story beats where ACTION/CAMERA/ENVIRONMENT change but visual identity stays consistent
4. Create detailed, cinematic prompts with structured components
5. Ensure each clip flows into the next through action progression, not just visual similarity

Return JSON with EXACTLY ${numClips} structured scenes:
{
  "scenes": [
    {
      "prompt": "Complete cinematic prompt for this story beat",
      "description": "What happens in this part of the story",
      "duration": "8-10 seconds",
      "sceneNumber": 1,
      "structure": {
        "camera": "Camera movement and angles",
        "lighting": "Lighting setup and mood",
        "subject": "Main subject and action",
        "mood": "Emotional tone for this beat",
        "pacing": "Timing and rhythm",
        "style": "Visual aesthetic",
        "transition": "How this clip connects to the next (e.g., 'Fade to similar angle', 'Match cut on motion')"
      }
    },
    ... (continue for all ${numClips} scenes)
  ]
}

VALIDATION CHECKLIST (Review before responding):
✓ Are the ${numClips} scenes showing DIFFERENT actions/moments (not variations of the same scene)?
✓ Does each scene have clear PROGRESSION in camera angle, subject action, or environment?
✓ Can you identify a clear beginning-middle-end arc across all clips?
✓ Would a viewer see these as a cohesive story rather than random variations?

CRITICAL: Return EXACTLY ${numClips} scenes with complete structure data and narrative continuity.`;
  }

  /**
   * Simple fallback parser if OpenAI is not available
   */
  static simpleParse(userPrompt, numClips = 1) {
    if (numClips === 1) {
      return [{
        prompt: userPrompt,
        description: 'User provided prompt',
        duration: '8-10 seconds'
      }];
    }

    // Simple multi-clip fallback: just split by periods or use same prompt
    return Array(numClips).fill(null).map((_, i) => ({
      prompt: userPrompt,
      description: `Scene ${i + 1}`,
      duration: '8-10 seconds',
      sceneNumber: i + 1
    }));
  }

  /**
   * Generate style guide based on creative options
   */
  getStyleGuide(options = {}) {
    const { style = 'cinematic', tone = 'premium', tempo = 'medium' } = options;

    let guide = `CREATIVE DIRECTION:\n`;

    // Style guidelines
    const styleMap = {
      'cinematic': '- Cinematic style: Use dramatic camera movements (dolly, crane shots), shallow depth of field, color grading with rich tones, professional composition following rule of thirds',
      'documentary': '- Documentary style: Handheld camera feel, natural lighting, authentic moments, candid angles, minimal color grading for realism',
      'energetic': '- Energetic style: Dynamic quick cuts implied through motion, bright vibrant colors, high contrast, fast-paced action, upbeat visual rhythm',
      'minimal': '- Minimal style: Clean compositions, negative space, simple backgrounds, muted color palette, elegant restraint, focus on essential elements',
      'dramatic': '- Dramatic style: High contrast lighting, bold shadows, intense moments, powerful angles (low/high), emotional close-ups, rich cinematic blacks',
      'playful': '- Playful style: Bright saturated colors, whimsical angles, creative framing, lighthearted energy, fun visual surprises'
    };
    guide += styleMap[style] || styleMap['cinematic'];
    guide += '\n';

    // Tone guidelines
    const toneMap = {
      'premium': '- Premium tone: Luxury aesthetics, refined details, sophisticated mood, high-end product treatment, aspirational feel',
      'friendly': '- Friendly tone: Warm approachable visuals, soft lighting, genuine smiles, welcoming environments, relatable scenarios',
      'edgy': '- Edgy tone: Bold unconventional angles, urban gritty textures, moody atmosphere, rebellious energy, modern attitude',
      'inspiring': '- Inspiring tone: Uplifting compositions, golden hour lighting when possible, triumphant moments, aspirational messaging, motivational energy',
      'humorous': '- Humorous tone: Unexpected visual gags, exaggerated expressions, lighthearted situations, comedic timing in action'
    };
    guide += toneMap[tone] || toneMap['premium'];
    guide += '\n';

    // Tempo guidelines
    const tempoMap = {
      'slow': '- Slow tempo: Deliberate pacing, lingering shots, gradual reveals, contemplative moments, smooth transitions, let scenes breathe',
      'medium': '- Medium tempo: Balanced pacing, natural rhythm, comfortable viewing pace, mix of wide and tight shots, steady progression',
      'fast': '- Fast tempo: Quick action, dynamic energy, rapid scene changes, high-energy subjects, punchy delivery, immediate impact'
    };
    guide += tempoMap[tempo] || tempoMap['medium'];

    return guide;
  }

  /**
   * Generate marketing framework based on AIDA and conversion psychology
   */
  getMarketingFramework(options = {}) {
    const { audience, goal = 'awareness', cta } = options;

    let framework = `MARKETING FRAMEWORK (AIDA - Attention, Interest, Desire, Action):\n`;

    // Audience targeting
    if (audience) {
      framework += `- Target Audience: ${audience}\n`;
      framework += `- Tailor visuals, pacing, and messaging to resonate with this specific demographic\n`;
      framework += `- Use relatable scenarios, environments, and emotional triggers relevant to ${audience}\n`;
    }

    // Goal-specific guidance
    const goalMap = {
      'awareness': '- Goal: Brand Awareness - Focus on memorable visuals, brand identity, and creating positive associations. Make it shareable and attention-grabbing',
      'sales': '- Goal: Drive Sales - Emphasize product benefits, urgency (limited time offers), social proof, and clear value propositions. Show transformation/results',
      'engagement': '- Goal: Boost Engagement - Create interactive, entertaining content that invites viewers to participate, comment, or share. Use hooks and intrigue',
      'signups': '- Goal: Generate Signups - Highlight exclusive benefits, ease of use, and what users gain. Remove friction, show simple steps'
    };
    framework += goalMap[goal] || goalMap['awareness'];
    framework += '\n';

    // AIDA structure
    framework += `\nAIDA Structure (apply across scenes):\n`;
    framework += `1. ATTENTION (Hook): Open with eye-catching visuals, unexpected moments, or bold statements (first 1-2 seconds are critical)\n`;
    framework += `2. INTEREST: Build curiosity through storytelling, show the problem or opportunity\n`;
    framework += `3. DESIRE: Evoke aspiration - show transformation, benefits, emotional payoff (use power words: exclusive, premium, revolutionary)\n`;
    framework += `4. ACTION: End with clear CTA${cta ? ` ("${cta}")` : ' (e.g., "Shop Now", "Learn More", "Join Today")'} - make it visible and compelling\n`;

    // Psychological triggers
    framework += `\nPsychological Triggers to Incorporate:\n`;
    framework += `- Urgency: "Limited time", "Now", "Today only" (if applicable)\n`;
    framework += `- Social Proof: Show people using/enjoying product, testimonials, crowds\n`;
    framework += `- Aspiration: Show ideal outcomes, lifestyle upgrade, success states\n`;
    framework += `- Contrast: Before/after, problem/solution dynamics\n`;

    return framework;
  }

  /**
   * Advanced cinematography guidance using professional film terminology
   */
  getCinematographyGuide(options = {}) {
    const { proCinematography } = options;

    if (!proCinematography) {
      return ''; // Skip if not enabled
    }

    return `
ADVANCED CINEMATOGRAPHY (Professional Film Terminology):

CAMERA MOVEMENTS (use precise terms, not vague descriptions):
- Dolly: Camera moves forward/backward on tracks (e.g., "Dolly in slowly from wide to medium close-up")
- Truck: Camera moves left/right parallel to subject (e.g., "Truck right following subject's walk")
- Pedestal: Camera moves up/down vertically (e.g., "Pedestal up revealing skyline")
- Crane/Boom: Sweeping vertical movements (e.g., "Crane shot rising from ground level to bird's eye")
- Pan: Camera rotates left/right on axis (e.g., "Slow pan left across product lineup")
- Tilt: Camera rotates up/down on axis (e.g., "Tilt down from face to hands holding product")
- Zoom: Lens focal length change (e.g., "Slow zoom in on subject's eyes, shallow depth of field")
- Handheld: Dynamic, energetic feel (e.g., "Handheld follow shot, slight camera shake for realism")
- Steadicam: Smooth tracking (e.g., "Steadicam tracking shot circling 360° around subject")
- Aerial/Drone: High vantage point (e.g., "Aerial drone shot descending toward location, revealing scale")

SHOT TYPES & FRAMING:
- Extreme Wide (EWS): Establishing shot, shows full environment
- Wide (WS): Full body + context
- Medium (MS): Waist up, conversational
- Close-Up (CU): Face/object detail, emotional
- Extreme Close-Up (ECU): Macro details (texture, eyes, product features)
- Over-the-shoulder (OTS): Perspective shots
- Dutch Angle: Tilted frame for tension/unease
- Rule of Thirds: Position subjects on intersecting lines for balanced composition

LIGHTING TECHNIQUES:
- Golden Hour: Warm, soft natural light (sunset/sunrise)
- Volumetric Lighting: God rays, atmospheric beams through windows/trees
- Rembrandt Lighting: Triangle of light on cheek, dramatic portrait style
- Backlighting: Subject lit from behind, rim light effect, silhouette or halo
- High-Key: Bright, minimal shadows (clean, optimistic)
- Low-Key: Dark, dramatic shadows (moody, mysterious)
- Practical Lights: Visible light sources in scene (lamps, neon signs)
- Color Temperature: Specify "warm 3200K tungsten glow" or "cool 5600K daylight"

VISUAL EFFECTS & TECHNIQUES:
- Lens Flares: Sun hitting lens for dreamy/epic feel
- Bokeh: Out-of-focus background lights creating circular blur
- Depth of Field: "Shallow DoF f/1.8 isolating subject" or "Deep DoF f/11 everything sharp"
- Motion Blur: Speed lines, fast action
- Slow Motion: 120fps or 240fps for dramatic emphasis
- Time-lapse: Compressed time showing change (clouds, crowds)
- Rack Focus: Shift focus from foreground to background mid-shot
- Particle Effects: Dust, sparkles, rain, snow for atmosphere
- Color Grading: "Teal and orange cinema grade" or "Desaturated Nordic noir palette"

TECHNICAL SPECS TO MENTION:
- Frame Rate: 24fps (cinematic), 30fps (broadcast), 60fps+ (smooth action)
- Aspect Ratio: 16:9 (standard), 9:16 (vertical/TikTok), 1:1 (square/Instagram), 2.39:1 (ultra-wide cinema)
- Film Stock References: "Kodak Vision3 5219 film look" or "ARRI Alexa warm sensor tone"

Example Professional Prompt:
"Open with aerial drone shot descending toward coastal highway at golden hour, 24fps. Dolly forward tracking sleek electric SUV, shallow depth of field f/2.8. Medium close-up of driver through windshield with lens flare from setting sun. Rack focus to rearview mirror reflection. Tilt down to wheels gripping asphalt. Crane up revealing ocean with volumetric god rays. End on wide shot, rule of thirds composition. Teal and orange color grade, cinematic film grain."
`;
  }

  /**
   * Platform-specific optimization guidance
   */
  getPlatformOptimization(options = {}) {
    const { platform = 'youtube' } = options;

    const platformMap = {
      'instagram': `
INSTAGRAM OPTIMIZATION:
- Aspect Ratio: 9:16 (Stories/Reels) or 1:1 (Feed posts)
- Hook: First 0.5 seconds must grab attention (platform favors watch time)
- Duration: 15-30 seconds ideal for Reels, 60 seconds max for feed
- Text Overlays: Use bold, readable fonts - many watch with sound off
- Visuals: Bright, high contrast, vibrant colors (mobile viewing)
- Pacing: Fast cuts, dynamic energy to prevent scrolling
- CTA: Place in first 3 seconds AND at end
- Audio: Trending sounds boost discoverability (mention if applicable)
`,
      'tiktok': `
TIKTOK OPTIMIZATION:
- Aspect Ratio: 9:16 (full vertical)
- Hook: First 1 second is CRITICAL - start with action, surprise, or bold statement
- Duration: 15-60 seconds (shorter often performs better)
- Native Feel: Handheld, authentic, less polished (avoid overly corporate)
- Text Overlays: Large, punchy text that's readable on small screens
- Trends: Reference trending formats, sounds, challenges if relevant
- Pacing: Very fast - new visual every 2-3 seconds
- CTA: Verbal + visual, natural integration (e.g., "Wait for it...")
`,
      'youtube': `
YOUTUBE OPTIMIZATION:
- Aspect Ratio: 16:9 (landscape)
- Hook: First 5 seconds prevent clicks away, establish value
- Duration: 30 seconds to 2 minutes for ads, longer for organic content
- Thumbnail Moment: Include a frame worth pausing on for thumbnail (high emotion, clear branding)
- Pacing: Moderate - build story with clear beginning, middle, end
- Production Quality: Higher polish expected (clean audio, stable footage)
- CTA: Clear verbal + visual CTA, can place mid-roll and end
- Branding: Logo/brand visible but not intrusive
`,
      'facebook': `
FACEBOOK OPTIMIZATION:
- Aspect Ratio: 1:1 (square) or 4:5 (vertical feed)
- Autoplay Silent: MUST work without sound - use captions/text overlays
- Hook: First 3 seconds shown in feed preview
- Duration: 15-30 seconds (attention span shorter on feed)
- Captions: Include full captions for accessibility and silent viewing
- Emotional Appeal: Facebook favors heartwarming, inspiring, or shocking content
- CTA: Clear button-style CTA graphic at end
- Shareability: Make it worth tagging friends or sharing
`
    };

    return platformMap[platform] || platformMap['youtube'];
  }
}
