import * as functions from 'firebase-functions/v1'; // v1 for auth triggers
import * as admin from 'firebase-admin';
import fetch from 'node-fetch';
import * as dotenv from 'dotenv';
import * as path from 'path';
import * as fs from 'fs';

// Load environment variables from root .env file
// Try multiple possible paths since the Firebase emulator changes working directories
const possibleEnvPaths = [
  path.join(__dirname, '../../.env'),
  path.join(process.cwd(), '.env'),
  path.join(__dirname, '../../../.env')
];

let envLoaded = false;
for (const envPath of possibleEnvPaths) {
  if (fs.existsSync(envPath)) {
    console.log(`Loading .env from: ${envPath}`);
    dotenv.config({ path: envPath });
    envLoaded = true;
    break;
  }
}

if (!envLoaded) {
  console.warn('No .env file found, using system environment variables');
}

// Debug: Log key environment variables
console.log('Environment check:', {
  BACKEND_URL: process.env.BACKEND_URL ? 'SET' : 'NOT SET',
  INTERNAL_API_KEY: process.env.INTERNAL_API_KEY ? 'SET' : 'NOT SET',
  NODE_ENV: process.env.NODE_ENV || 'undefined'
});

// Initialize Firebase Admin SDK
admin.initializeApp();

interface SyncUserRequest {
  firebase_uid: string;
  email: string;
  display_name: string;
  photo_url: string;
}

interface SyncUserResponse {
  name: string;
  first_name: string;
  last_name: string;
  email: string;
  avatar: string;
}

/**
 * Cloud Function triggered when a Firebase user is created.
 * Syncs the user data to the internal PostgreSQL database via Connect-RPC API.
 */
export const syncUserOnCreate = functions.auth.user().onCreate(async (user: admin.auth.UserRecord) => {
  const functionName = 'syncUserOnCreate';
  const startTime = Date.now();

  console.log(`${functionName}: Starting user sync for Firebase UID: ${user.uid}`);

  try {
    // Validate required environment variables
    const backendUrl = process.env.BACKEND_URL;
    const internalApiKey = process.env.INTERNAL_API_KEY;

    if (!backendUrl) {
      throw new Error('BACKEND_URL environment variable is not set');
    }

    if (!internalApiKey) {
      throw new Error('INTERNAL_API_KEY environment variable is not set');
    }

    // Prepare sync request payload
    const syncRequest: SyncUserRequest = {
      firebase_uid: user.uid,
      email: user.email || '',
      display_name: user.displayName || '',
      photo_url: user.photoURL || ''
    };

    console.log(`${functionName}: Calling API endpoint: ${backendUrl}/pb.ArtGeneratorService/SyncUserFromFirebase`);
    console.log(`${functionName}: Request payload:`, JSON.stringify(syncRequest, null, 2));

    // Make HTTP request to Connect-RPC API
    const response = await fetch(`${backendUrl}/pb.ArtGeneratorService/SyncUserFromFirebase`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${internalApiKey}`,
      },
      body: JSON.stringify(syncRequest),
      timeout: 10000, // 10 second timeout
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`API call failed: ${response.status} ${response.statusText}. Response: ${errorText}`);
    }

    const syncResponse: SyncUserResponse = await response.json();
    const duration = Date.now() - startTime;

    console.log(`${functionName}: User synced successfully in ${duration}ms`);
    console.log(`${functionName}: Created/Retrieved user:`, JSON.stringify(syncResponse, null, 2));

    return { 
      success: true, 
      user_id: syncResponse.name,
      duration_ms: duration 
    };

  } catch (error) {
    const duration = Date.now() - startTime;
    
    console.error(`${functionName}: Failed to sync user after ${duration}ms:`, error);
    
    // Log additional context for debugging
    console.error(`${functionName}: User data:`, {
      uid: user.uid,
      email: user.email,
      displayName: user.displayName,
      photoURL: user.photoURL,
      providerData: user.providerData
    });

    // Re-throw error to trigger Cloud Function retry mechanism
    throw error;
  }
});

/**
 * Health check endpoint for monitoring Cloud Functions
 */
export const healthCheck = functions.https.onRequest((request: functions.https.Request, response: functions.Response) => {
  response.json({
    status: 'healthy',
    timestamp: new Date().toISOString(),
    environment: process.env.NODE_ENV || 'development'
  });
});