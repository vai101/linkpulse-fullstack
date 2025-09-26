import { useState, useEffect, useCallback } from 'react';
import './App.css';

function App() {
  const [analytics, setAnalytics] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);
  const [longUrl, setLongUrl] = useState('');
  const [newShortUrl, setNewShortUrl] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  // New state for the copy button
  const [copied, setCopied] = useState(false);

  const API_URL = import.meta.env.VITE_API_URL;
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  const fetchAnalytics = useCallback(() => {
    fetch(API_URL)
      .then(response => response.json())
      .then(data => {
        setAnalytics(data);
        setIsLoading(false);
      })
      .catch(error => {
        setError(error.message);
        setIsLoading(false);
      });
  }, [API_URL]);

  useEffect(() => {
    fetchAnalytics();
  }, [fetchAnalytics]);

  const handleCopy = () => {
    navigator.clipboard.writeText(newShortUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000); // Reset after 2 seconds
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    if (!longUrl) return;

    setIsSubmitting(true);
    setNewShortUrl('');
    setError(null);

    try {
      const response = await fetch(API_BASE_URL + '/', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url: longUrl }),
      });

      if (!response.ok) {
        throw new Error('Failed to create short URL');
      }

      const data = await response.json();
      setNewShortUrl(data.short_url);
      setLongUrl('');
      fetchAnalytics();
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
      <header>
        <h1>LinkPulse</h1>
        <p className="subtitle">A simple, fast, and elegant URL shortener.</p>
      </header>
      
      <main>
        <form onSubmit={handleSubmit} className="url-form">
          <input
            type="url"
            placeholder="https://your-long-url.com/goes-here"
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
            <a href={newShortUrl} target="_blank" rel="noopener noreferrer">
              {newShortUrl}
            </a>
            <button onClick={handleCopy} className="copy-button">
              {copied ? 'Copied!' : 'Copy'}
            </button>
          </div>
        )}

        <h2>Analytics</h2>
        {isLoading ? (
          <p>Loading analytics...</p>
        ) : (
          <div className="table-container">
            <table>
              <thead>
                <tr>
                  <th>Short URL</th>
                  <th>Original URL</th>
                  <th>Clicks</th>
                </tr>
              </thead>
              <tbody>
                {analytics.map(item => (
                  <tr key={item.short_code}>
                    <td>
                      <a href={`${API_BASE_URL}/${item.short_code}`} target="_blank" rel="noopener noreferrer">
                        {`${API_BASE_URL.replace('https://','')}/${item.short_code}`}
                      </a>
                    </td>
                    <td className="long-url" title={item.long_url}>
                      <span>{item.long_url}</span>
                    </td>
                    <td>{item.click_count}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </main>
    </div>
  );
}

export default App;