# Audio Quality & Synchronization Manual Test

## Test Environment Setup

- [ ] Use a generated 30s pharmaceutical ad with all four tracks populated.
- [ ] Confirm background music and narrator audio URLs are valid presigned links.
- [ ] Prepare devices:
  - Desktop: Chrome (latest), Safari (latest), Firefox (latest)
  - Mobile: iOS Safari, Android Chrome
- [ ] Prepare output hardware:
  - Built-in laptop speakers
  - Wired or Bluetooth headphones
  - External speakers (if available)
- [ ] Clear browser cache to avoid cached audio artifacts.

## Volume Level Validation

1. Set workspace volume slider to **0%**
   - [ ] Music silent
   - [ ] Narrator silent
   - [ ] Video remains muted (expected)
2. Set volume slider to **50%**
   - [ ] Background music ≈15% actual volume (0.3 × 0.5)
   - [ ] Narrator ≈50% actual volume (1.0 × 0.5)
   - [ ] Narrator remains intelligible over music
3. Set volume slider to **100%**
   - [ ] Background music at 30% base level
   - [ ] Narrator at full level (clear and dominant)
   - [ ] No clipping or distortion at max volume

## Playback & Sync Verification

1. **Initial Play**
   - [ ] Click play → video, music, narrator start within 50ms
2. **Pause / Resume**
   - [ ] Pause → all tracks halt simultaneously
   - [ ] Resume → all tracks resume in sync
3. **Seek Operations**
   - [ ] Seek to midpoint (e.g., 15s) → music & narrator align instantly
   - [ ] Seek to last 5s → side effects audio aligns with overlay
4. **Long Playback**
   - [ ] Let video play full duration
   - [ ] Watch console logs for drift warnings
   - [ ] Confirm no drift exceeds ±100ms
5. **Drift Handling**
   - [ ] Observe automatic resync if drift > 200ms (`[VIDEO_PLAYER]` log)
   - [ ] Observe toast warning only when drift > 500ms

## Side Effects Segment Testing

- [ ] Side effects text overlay appears only in final 20% of timeline.
- [ ] Narrator speeds up to 1.4× in final segment (fast yet intelligible).
- [ ] Background music maintains constant tempo.
- [ ] Overall blend remains compliant (narrator clearly audible).

## Device Coverage

### Desktop Browsers

- [ ] **Chrome**: Sync stable, no autoplay blocking after user interaction.
- [ ] **Safari**: Audio tracks honor volume mix, no drift > ±100ms.
- [ ] **Firefox**: Console free of media errors, sync maintained.

### Mobile Browsers

- [ ] **iOS Safari**: Audio starts after user gesture, stays synchronized.
- [ ] **Android Chrome**: Playback consistent, no significant buffering drift.

### Output Hardware

- [ ] **Headphones**: Narrator centered, music ambient, no hiss.
- [ ] **Laptop Speakers**: Speech intelligible, music not overpowering.
- [ ] **External Speakers**: Levels consistent with other outputs.

## Audio Integrity Checks (Optional)

Run on exported audio assets to confirm quality:

```bash
# Replace paths with actual presigned URLs or downloaded assets
ffmpeg -i background-music.mp3 -af "astats" -f null - 2>&1 | grep 'Peak level'
ffmpeg -i narrator-voiceover.mp3 -af "astats" -f null - 2>&1 | grep 'Peak level'
```

- [ ] Peak levels stay below -0.1 dBFS (no clipping).
- [ ] RMS levels within expected range (music lower than narrator).

## Pass / Fail Criteria

- All checklist items completed on at least one desktop and one mobile device.
- No persistent drift beyond ±100ms after auto-resync.
- Narrator clearly intelligible over music at default settings.
- Side effects narration intelligible at 1.4× speed.
- No clipping, crackling, or distortion detected in either track.

Document any deviations, including device, browser, timestamp, and reproduction steps. Open follow-up tickets for issues that block compliance.
