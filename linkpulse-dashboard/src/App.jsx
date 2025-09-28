import { useState, useEffect } from 'react';
import './App.css';

function App() {
  const [analytics, setAnalytics] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);
  const [longUrl, setLongUrl] = useState('');
  const [newShortUrl, setNewShortUrl] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [copied, setCopied] = useState(false);

  const API_URL = import.meta.env.VITE_API_URL;
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  // This useEffect hook will run only once when the component mounts
  useEffect(() => {
    const fetchData = () => {
      // We add a timestamp to every request to bypass any possible cache
      const urlWithCacheBusting = `${API_URL}?t=${new Date().getTime()}`;
      
      fetch(urlWithCacheBusting, { method: 'POST' })
        .then(response => {
          if (!response.ok) {
            throw new Error('Network response was not ok');
          }
          return response.json();
        })
        .then(data => {
          setAnalytics(data);
        })
        .catch(error => {
          // Don't set an error for polling fails, just log it
          console.error("Failed to fetch analytics:", error);
        })
        .finally(() => {
          setIsLoading(false); // Only matters on the first load
        });
    };

    fetchData(); // Fetch data immediately on load
    const intervalId = setInterval(fetchData, 5000); // And then fetch every 5 seconds

    // This cleanup function is crucial to stop the polling when you leave the page
    return () => clearInterval(intervalId);
  }, [API_URL]);


  const handleCopy = () => {
    navigator.clipboard.writeText(newShortUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
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
      // Manually trigger a fetch right after success
      const updatedAnalytics = await (await fetch(API_URL, { method: 'POST' })).json();
      setAnalytics(updatedAnalytics);
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