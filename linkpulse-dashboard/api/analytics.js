// linkpulse-dashboard/api/analytics.js

export default async function handler(request, response) {
  // Get the real API URL from the environment variables we set in Vercel
  const renderApiUrl = process.env.VITE_API_URL;

  try {
    // Make a POST request from the serverless function to our Render API
    const apiResponse = await fetch(renderApiUrl, {
      method: 'POST',
    });

    if (!apiResponse.ok) {
      throw new Error(`API request failed with status ${apiResponse.status}`);
    }

    const data = await apiResponse.json();

    // Important: Tell Vercel not to cache the result of this function
    response.setHeader('Cache-Control', 'no-cache, no-store, must-revalidate');

    // Send the fresh data back to the browser
    response.status(200).json(data);

  } catch (error) {
    response.status(500).json({ error: error.message });
  }
}