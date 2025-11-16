// Placeholder Lambda function for Script Parser
// This is a temporary implementation that returns mock data
// Replace with actual Go binary when ready

exports.handler = async (event) => {
    console.log('Parser Lambda invoked with event:', JSON.stringify(event));

    const { prompt, duration = 30, user_id } = event;

    // Mock script generation
    const scriptId = `script-${Date.now()}`;

    return {
        statusCode: 200,
        script_id: scriptId,
        user_id: user_id,
        title: 'Mock Generated Script',
        scenes: [
            {
                scene_number: 1,
                start_time: 0,
                duration: duration / 2,
                location: 'INT. SCENE - DAY',
                action: 'Mock scene 1 placeholder',
                generation_prompt: prompt
            },
            {
                scene_number: 2,
                start_time: duration / 2,
                duration: duration / 2,
                location: 'EXT. SCENE - DAY',
                action: 'Mock scene 2 placeholder',
                generation_prompt: prompt
            }
        ],
        status: 'draft',
        message: 'Placeholder: Script parsed successfully (mock data)'
    };
};
