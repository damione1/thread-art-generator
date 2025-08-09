import * as functions from 'firebase-functions/v1'; // v1 for auth triggers
import * as admin from 'firebase-admin';
import fetch, { AbortError } from 'node-fetch';
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
  INTERNAL_API_KEY_VALUE: process.env.INTERNAL_API_KEY, // Temporary debug logging
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
 * Enhanced with better error handling and retry mechanisms.
 */
export const syncUserOnCreate = functions.auth.user().onCreate(async (user: admin.auth.UserRecord) => {
  const functionName = 'syncUserOnCreate';
  const startTime = Date.now();

  console.log(`${functionName}: Starting user sync for Firebase UID: ${user.uid}`);

  try {
    // Validate required environment variables with detailed error messages
    const backendUrl = process.env.BACKEND_URL;
    const internalApiKey = process.env.INTERNAL_API_KEY;

    if (!backendUrl) {
      const error = new Error(`${functionName}: BACKEND_URL environment variable is not set. Cannot proceed with user sync.`);
      console.error(error.message);
      throw error;
    }

    if (!internalApiKey) {
      const error = new Error(`${functionName}: INTERNAL_API_KEY environment variable is not set. Cannot authenticate with backend.`);
      console.error(error.message);
      throw error;
    }

    console.log(`${functionName}: Environment validated - Backend URL: ${backendUrl.substring(0, 30)}...`);

    // Add delay to prevent race conditions with auth events
    if (process.env.FUNCTIONS_EMULATOR === 'true') {
      console.log(`${functionName}: Running in emulator mode - adding delay to prevent race conditions`);
      await new Promise(resolve => setTimeout(resolve, 500));
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

    // Retry mechanism for backend API calls
    const maxRetries = 3;
    let lastError: Error | undefined;
    let syncResponse: SyncUserResponse | null = null;

    for (let attempt = 1; attempt <= maxRetries; attempt++) {
      try {
        console.log(`${functionName}: API call attempt ${attempt}/${maxRetries}`);

        // Make HTTP request to Connect-RPC API with timeout
        // Note: SyncUserFromFirebase requires internal API key validation (separate from auth interceptor)
        
        // Create timeout promise
        const timeoutPromise = new Promise((_, reject) => {
          setTimeout(() => {
            reject(new Error('Request timeout after 10 seconds'));
          }, 10000);
        });

        // Create fetch promise
        const fetchPromise = fetch(`${backendUrl}/pb.ArtGeneratorService/SyncUserFromFirebase`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Connect-Protocol-Version': '1',
            'Accept': 'application/json',
            'Authorization': `Bearer ${internalApiKey}`, // Required for internal API key validation
          },
          body: JSON.stringify(syncRequest),
        });

        const response = await Promise.race([fetchPromise, timeoutPromise]) as any;

        if (!response.ok) {
          const errorText = await response.text();
          const errorMessage = `API call failed: ${response.status} ${response.statusText}. Response: ${errorText}`;
          
          // Don't retry on client errors (4xx)
          if (response.status >= 400 && response.status < 500) {
            throw new Error(`${functionName}: ${errorMessage} (Client error - not retrying)`);
          }
          
          throw new Error(`${functionName}: ${errorMessage}`);
        }

        syncResponse = await response.json();
        console.log(`${functionName}: API call successful on attempt ${attempt}`);
        break; // Success - exit retry loop

      } catch (error) {
        const errorObj = error instanceof Error ? error : new Error(String(error));
        lastError = errorObj;
        console.warn(`${functionName}: API call attempt ${attempt} failed:`, error);

        // Check if it's an AbortError (timeout)
        if (error instanceof AbortError || errorObj.name === 'AbortError') {
          console.warn(`${functionName}: Request timed out on attempt ${attempt}`);
        }

        // If this is the last attempt, we'll throw the error
        if (attempt === maxRetries) {
          console.error(`${functionName}: All ${maxRetries} API call attempts failed`);
          break;
        }

        // Exponential backoff delay before retry
        const delay = Math.min(1000 * Math.pow(2, attempt - 1), 5000); // Max 5 seconds
        console.log(`${functionName}: Retrying in ${delay}ms...`);
        await new Promise(resolve => setTimeout(resolve, delay));
      }
    }

    if (!syncResponse) {
      throw lastError || new Error(`${functionName}: Failed to sync user after ${maxRetries} attempts`);
    }
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
    const errorMessage = error instanceof Error ? error.message : String(error);

    console.error(`${functionName}: Failed to sync user after ${duration}ms:`, error);

    // Enhanced error logging with categorization
    const errorCategory = errorMessage.includes('BACKEND_URL') || errorMessage.includes('INTERNAL_API_KEY') 
      ? 'CONFIGURATION_ERROR'
      : errorMessage.includes('Client error - not retrying')
      ? 'CLIENT_ERROR'
      : errorMessage.includes('timed out')
      ? 'TIMEOUT_ERROR'
      : 'UNKNOWN_ERROR';

    console.error(`${functionName}: Error category: ${errorCategory}`);

    // Log additional context for debugging
    console.error(`${functionName}: User data:`, {
      uid: user.uid,
      email: user.email,
      displayName: user.displayName,
      photoURL: user.photoURL,
      providerData: user.providerData.map(p => ({
        providerId: p.providerId,
        uid: p.uid,
        email: p.email,
        displayName: p.displayName
      })),
      metadata: {
        creationTime: user.metadata.creationTime,
        lastSignInTime: user.metadata.lastSignInTime
      }
    });

    // Log environment context
    console.error(`${functionName}: Environment context:`, {
      nodeEnv: process.env.NODE_ENV,
      functionsEmulator: process.env.FUNCTIONS_EMULATOR,
      backendUrlSet: !!process.env.BACKEND_URL,
      internalApiKeySet: !!process.env.INTERNAL_API_KEY,
      timestamp: new Date().toISOString()
    });

    // For configuration errors, don't retry (they won't succeed)
    if (errorCategory === 'CONFIGURATION_ERROR') {
      console.error(`${functionName}: Configuration error detected - not retrying`);
    } else {
      console.error(`${functionName}: Retryable error - Cloud Functions will attempt retry`);
    }

    // Re-throw error to trigger Cloud Function retry mechanism (except for config errors)
    throw error;
  }
});

/**
 * Cloud Function triggered when a file is uploaded to Firebase Storage.
 * Processes art image uploads and updates art status via Connect-RPC API.
 * 
 * The storage path format is: users/{internal_user_id}/arts/{art_id}
 * This matches the Google AIP resource name format used by the API.
 */
export const onArtImageUpload = functions.storage.object().onFinalize(async (object: functions.storage.ObjectMetadata) => {
  const functionName = 'onArtImageUpload';
  const startTime = Date.now();

  console.log(`${functionName}: Processing storage upload for: ${object.name}`);

  try {
    // Only process files matching the expected art upload path pattern
    if (!object.name || !object.name.match(/^users\/[^\/]+\/arts\/[^\/]+$/)) {
      console.log(`${functionName}: Ignoring file - not matching art upload pattern: ${object.name}`);
      return null;
    }

    // Validate required environment variables
    const backendUrl = process.env.BACKEND_URL;
    const internalApiKey = process.env.INTERNAL_API_KEY;

    if (!backendUrl) {
      throw new Error('BACKEND_URL environment variable is not set');
    }

    if (!internalApiKey) {
      throw new Error('INTERNAL_API_KEY environment variable is not set');
    }

    // Determine environment context
    const isEmulator = process.env.FUNCTIONS_EMULATOR === 'true' ||
      process.env.FIREBASE_AUTH_EMULATOR_HOST ||
      object.bucket?.includes('demo-');

    console.log(`${functionName}: Environment: ${isEmulator ? 'emulator' : 'production'}`);

    // Build image URL based on environment
    let imageUrl: string;
    
    if (isEmulator) {
      const storageEmulatorHost = process.env.FIREBASE_STORAGE_EMULATOR_EXTERNAL_HOST || 'localhost:9199';
      const bucketName = object.bucket || 'demo-thread-art-generator.appspot.com';
      imageUrl = `http://${storageEmulatorHost}/v0/b/${bucketName}/o/${encodeURIComponent(object.name!)}?alt=media`;
    } else {
      // Production: Use signed URL with Admin SDK
      const bucket = admin.storage().bucket(object.bucket);
      const file = bucket.file(object.name);
      const [signedUrl] = await file.getSignedUrl({
        action: 'read',
        expires: '03-01-2500',
      });
      imageUrl = signedUrl;
    }

    console.log(`${functionName}: Generated image URL: ${imageUrl.substring(0, 80)}...`);

    // Extract Firebase UID from the path
    // The storage path format is: users/{firebase_uid}/arts/{art_id}
    const pathParts = object.name.split('/');
    const firebaseUidFromPath = pathParts[1]; // This should be Firebase UID
    
    // In production, validate Firebase UID from metadata if available
    let firebaseUid = firebaseUidFromPath;
    
    if (!isEmulator && object.metadata?.uploadedBy) {
      // In production, verify the Firebase UID from metadata matches the path
      if (object.metadata.uploadedBy !== firebaseUidFromPath) {
        throw new Error(`Firebase UID mismatch: path=${firebaseUidFromPath}, metadata=${object.metadata.uploadedBy}`);
      }
      console.log(`${functionName}: Validated Firebase UID consistency between path and metadata`);
    }
    
    if (!firebaseUid) {
      throw new Error('Firebase UID not found in storage path');
    }

    // The resource name in the request is the same as the storage path
    const resourceName = object.name;
    
    const confirmRequest = {
      name: resourceName, // This is the key insight - resource name = storage path
      firebase_uid: firebaseUid,
      image_url: imageUrl,
      original_filename: object.name.split('/').pop() || 'uploaded_image',
      content_type: object.contentType === 'application/octet-stream' ? '' : (object.contentType || ''),
      file_size: parseInt(object.size || '0')
    };

    console.log(`${functionName}: Confirming art upload:`, {
      resource_name: confirmRequest.name,
      firebase_uid: confirmRequest.firebase_uid,
      content_type: confirmRequest.content_type,
      file_size: confirmRequest.file_size
    });

    // Call backend API to confirm upload
    const response = await fetch(`${backendUrl}/pb.FirebaseFunctionsService/ConfirmArtImageUploadFromFunction`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Connect-Protocol-Version': '1',
        'Authorization': `Bearer ${internalApiKey}`,
      },
      body: JSON.stringify(confirmRequest),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Backend API call failed: ${response.status} ${response.statusText}. Response: ${errorText}`);
    }

    const result = await response.json();
    const duration = Date.now() - startTime;

    console.log(`${functionName}: Successfully processed art upload in ${duration}ms`);
    console.log(`${functionName}: Art status updated:`, {
      art_id: result.id,
      status: result.status,
      image_url: result.image_url ? 'SET' : 'NOT SET'
    });

    return {
      success: true,
      resource_name: resourceName,
      duration_ms: duration
    };

  } catch (error) {
    const duration = Date.now() - startTime;
    console.error(`${functionName}: Failed to process upload after ${duration}ms:`, error);
    
    // Log object details for debugging
    console.error(`${functionName}: Object details:`, {
      name: object.name,
      bucket: object.bucket,
      contentType: object.contentType,
      size: object.size,
      metadata: object.metadata
    });

    throw error;
  }
});

/**
 * Health check endpoint for monitoring Cloud Functions
 */
export const healthCheck = functions.https.onRequest((_request: functions.https.Request, response: functions.Response) => {
  response.json({
    status: 'healthy',
    timestamp: new Date().toISOString(),
    environment: process.env.NODE_ENV || 'development'
  });
});