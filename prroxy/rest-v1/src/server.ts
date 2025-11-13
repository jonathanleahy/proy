import { createApp } from './app';

const PORT = process.env.PORT || 3000;

const app = createApp();

app.listen(PORT, () => {
  console.log(`🚀 REST API v1 server running on http://0.0.0.0:${PORT}`);
  console.log(`📊 Health check: http://0.0.0.0:${PORT}/health`);
  console.log('\nAvailable endpoints:');
  console.log(`  GET  /api/user/:id - Get user data`);
  console.log(`  GET  /api/user/:id/summary - Get user summary with posts`);
  console.log(`  POST /api/user/:id/report - Get comprehensive user report`);
  console.log(`  GET  /api/person?surname=X&dob=YYYY-MM-DD - Find person by surname and DOB`);
});
