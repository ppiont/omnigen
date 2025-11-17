import fs from 'fs';
import path from 'path';

/**
 * Export scene data to structured XML format for editing
 */
export class XMLExporter {
  constructor(outputDir = './output') {
    this.outputDir = path.resolve(outputDir);
    this.ensureOutputDir();
  }

  ensureOutputDir() {
    if (!fs.existsSync(this.outputDir)) {
      fs.mkdirSync(this.outputDir, { recursive: true });
      console.log(`📁 Created output directory: ${this.outputDir}`);
    }
  }

  /**
   * Convert scene object to XML
   */
  sceneToXML(scene, index) {
    const structure = scene.structure || {};

    return `  <clip id="${index + 1}" scene_number="${scene.sceneNumber || index + 1}">
    <description>${this.escapeXML(scene.description || '')}</description>
    <duration>${scene.duration || '8-10 seconds'}</duration>

    <prompt>${this.escapeXML(scene.prompt)}</prompt>

    <structure>
      <camera>${this.escapeXML(structure.camera || 'Not specified')}</camera>
      <lighting>${this.escapeXML(structure.lighting || 'Not specified')}</lighting>
      <subject>${this.escapeXML(structure.subject || 'Not specified')}</subject>
      <mood>${this.escapeXML(structure.mood || 'Not specified')}</mood>
      <pacing>${this.escapeXML(structure.pacing || 'Not specified')}</pacing>
      <style>${this.escapeXML(structure.style || 'Not specified')}</style>
      ${structure.transition ? `<transition>${this.escapeXML(structure.transition)}</transition>` : ''}
    </structure>
  </clip>`;
  }

  /**
   * Escape special XML characters
   */
  escapeXML(str) {
    if (!str) return '';
    return str
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&apos;');
  }

  /**
   * Export scenes to XML file
   */
  exportToXML(scenes, originalPrompt, metadata = {}) {
    const timestamp = new Date().toISOString();
    const xmlContent = `<?xml version="1.0" encoding="UTF-8"?>
<video_sequence>
  <metadata>
    <generated_at>${timestamp}</generated_at>
    <original_prompt>${this.escapeXML(originalPrompt)}</original_prompt>
    <total_clips>${scenes.length}</total_clips>
    ${metadata.model ? `<model>${this.escapeXML(metadata.model)}</model>` : ''}
    ${metadata.total_duration ? `<total_duration>${metadata.total_duration}</total_duration>` : ''}
  </metadata>

  <clips>
${scenes.map((scene, i) => this.sceneToXML(scene, i)).join('\n\n')}
  </clips>
</video_sequence>
`;

    const timestamp_file = new Date().toISOString().replace(/[:.]/g, '-');
    const xmlPath = path.join(this.outputDir, `scene_structure_${timestamp_file}.xml`);

    fs.writeFileSync(xmlPath, xmlContent, 'utf-8');
    console.log(`📄 XML structure exported to: ${xmlPath}`);

    return xmlPath;
  }

  /**
   * Import XML and parse back to scenes
   * (For future editing workflow)
   */
  importFromXML(xmlPath) {
    // TODO: Implement XML parsing to reload scenes for editing
    // This would allow users to edit the XML and regenerate specific clips
    console.log('XML import feature - coming soon!');
  }
}
