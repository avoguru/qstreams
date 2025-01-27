// Page Switching Logic
document.getElementById('dashboard-link').addEventListener('click', function () {
    setActivePage('dashboard-page', this);
});
document.getElementById('api-link').addEventListener('click', function () {
    setActivePage('api-page', this);
});

function setActivePage(pageId, linkElement) {
    document.querySelectorAll('.page').forEach(page => page.classList.remove('active'));
    document.getElementById(pageId).classList.add('active');
    
    document.querySelectorAll('.menu li a').forEach(link => link.classList.remove('active'));
    linkElement.classList.add('active');
}

// Fetch Streams and Metrics and Update the Dashboard
async function fetchStreams() {
    try {
        const streamsResponse = await fetch('/streams');
        const streamsData = await streamsResponse.json();

        const metricsResponse = await fetch('/metrics');
        const metricsData = await metricsResponse.json();

        const streamTiles = document.getElementById('stream-tiles');
        streamTiles.innerHTML = ''; // Clear existing tiles

        // Populate the dashboard with updated streams and metrics
        streamsData.streams.forEach(stream => {
            const metrics = metricsData.streams.find(m => m.stream_id === stream.stream_id) || {};
            const tile = document.createElement('div');
            tile.className = 'tile';

            tile.innerHTML = `
                <h2>${stream.name}</h2>
                <p><strong>ID:</strong> ${stream.stream_id}</p>
                <p><strong>Events Sent:</strong> ${metrics.events_sent || 0}</p>
                <p><strong>Events Deduped:</strong> ${metrics.events_deduped || 0}</p>
                <p><strong>Queries:</strong> ${metrics.number_of_queries || 0}</p>
            `;

            streamTiles.appendChild(tile);
        });
    } catch (error) {
        console.error('Error fetching streams or metrics:', error);
    }
}

// Auto-refresh Metrics Every 10 Seconds
setInterval(fetchStreams, 2000);

// Fetch Streams on Page Load
fetchStreams();