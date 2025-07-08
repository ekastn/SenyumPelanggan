const express = require('express');
const path = require('path');

const app = express();
const port = process.env.PORT || 3000;

// Endpoint to provide backend API URL to the frontend
app.get('/config', (req, res) => {
  res.json({ backendApiUrl: process.env.BACKEND_API_URL || 'http://localhost:8080' });
});

// Serve static files from the current directory
app.use(express.static(__dirname));

app.listen(port, () => {
  console.log(`Frontend server listening at http://localhost:${port}`);
});
