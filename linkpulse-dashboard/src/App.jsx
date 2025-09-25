// src/App.jsx

import { useState, useEffect, useCallback } from 'react';
import './App.css';

function App() {
  const [analytics, setAnalytics] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  // State for our new form
  const [longUrl, setLongUrl] = useState('');
  const [newShortUrl, setNewShortUrl] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  // We wrap the fetch logic in a useCallback so we can reuse it
  const fetchAnalytics = useCallback(() => {
    fetch(import.meta.env.VITE_API_URL)
      .then(response => response.json())
      .then(data => {
        setAnalytics(data);
        setIsLoading(false);
      })
      .catch(error => {
        setError(error.message);
        setIsLoading(false);
      });
  }, []);

  // Run the fetch logic when the component first loads
  useEffect(() => {
    fetchAnalytics();
  }, [fetchAnalytics]);

  // This function handles the form submission
  const handleSubmit = async (event) => {
    event.preventDefault(); // Prevent the browser from reloading the page
    if (!longUrl) return; // Don't submit if the input is empty

    setIsSubmitting(true);
    setNewShortUrl('');

    try {
      const response = await fetch('http://localhost:8080/', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ url: longUrl }),
      });

      if (!response.ok) {
        throw new Error('Failed to create short URL');
      }

      const data = await response.json();
      setNewShortUrl(data.short_url); // Save the new short URL to display it
      setLongUrl(''); // Clear the input field
      fetchAnalytics(); // Refresh the analytics table
    } catch (err) {
      setError(err.message);
    } finally {
      setIsSubmitting(false);
    }
  };

  if (error) {
    return <div className="container">Error: {error}</div>;
  }

  return (
    <div className="container">
      <h1>LinkPulse</h1>
      
      <form onSubmit={handleSubmit} className="url-form">
        <input
          type="url"
          placeholder="Enter a long URL to shorten"
          value={longUrl}
          onChange={(e) => setLongUrl(e.target.value)}
          required
        />
        <button type="submit" disabled={isSubmitting}>
          {isSubmitting ? 'Shortening...' : 'Shorten'}
        </button>
      </form>

      {newShortUrl && (
        <div className="result">
          <p>Your short URL:</p>
          <a href={newShortUrl} target="_blank" rel="noopener noreferrer">
            {newShortUrl}
          </a>
        </div>
      )}

      <h2>Analytics</h2>
      {isLoading ? (
        <p>Loading analytics...</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>Short Code</th>
              <th>Original URL</th>
              <th>Clicks</th>
            </tr>
          </thead>
          <tbody>
            {analytics.map(item => (
              <tr key={item.short_code}>
                <td>
                  <a href={`http://localhost:8080/${item.short_code}`} target="_blank" rel="noopener noreferrer">
                    {item.short_code}
                  </a>
                </td>
                <td className="long-url">{item.long_url}</td>
                <td>{item.click_count}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}

export default App;