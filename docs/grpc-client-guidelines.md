# Thread Art Generator - gRPC Client Implementation Guidelines

This document provides guidelines for working with the gRPC client in the Thread Art Generator application.

## Architecture Overview

Our system uses:

1. Next.js frontend connecting to a gRPC backend via gRPC-Web
2. Envoy proxy to handle HTTP/gRPC protocol translation
3. Token-based authentication with automatic refresh

The request flow is as follows:

- Frontend makes gRPC-Web calls to backend services
- Requests go through Envoy proxy (port 443), which forwards them to the appropriate backend service
- Authentication is handled via Bearer tokens with automatic refresh logic

## Core Architecture Components

### 1. gRPC Client Setup

```typescript
// Client creation with transport configuration
const transport = createGrpcWebTransport({
  baseUrl: CONFIG.baseUrl,
  useBinaryFormat: true,
  credentials: "include",
});
const client = createClient(ArtGeneratorService, transport);
```

### 2. Authentication Handling

Our authentication system includes:

- Token caching to minimize requests
- Automatic token refresh when expired
- Authorization header injection
- Retry mechanism for auth failures

```typescript
// Example of token handling
export const getAccessToken = async (): Promise<string> => {
  // Check if we have a cached token that's not expired
  if (
    tokenCache.token &&
    Date.now() < tokenCache.expiresAt - CONFIG.tokenExpiryBufferMs
  ) {
    return tokenCache.token;
  }

  // Fetch a new token
  return fetchAccessToken();
};
```

### 3. Service Wrapper

The `GrpcService` class provides:

- Standardized error handling
- Automatic token refresh on auth errors
- Simplified method calling pattern

```typescript
export class GrpcService {
  static async call<T>(
    serviceCall: (token: string | undefined) => Promise<T>,
    forceFetchToken = false
  ): Promise<T> {
    try {
      // Get the current token (or force fetch a new one)
      const token = forceFetchToken
        ? await fetchAccessToken()
        : await getAccessToken();

      // Make the call with the token
      return await serviceCall(token);
    } catch (error) {
      // Handle auth errors by refreshing the token and retrying once
      if (
        error instanceof ConnectError &&
        (error.code === Code.Unauthenticated ||
          error.code === Code.PermissionDenied) &&
        !forceFetchToken
      ) {
        // Try one more time with a fresh token
        return GrpcService.call(serviceCall, true);
      }

      // Otherwise rethrow
      throw error;
    }
  }
}
```

## Guidelines for Implementing Endpoints

### Adding New Endpoints

When adding new endpoints, follow this pattern:

```typescript
export const someNewEndpoint = async (params) => {
  // Dynamic import of the proto definition
  const { SomeNewRequest } = await import("./pb/some_file_pb");

  return GrpcService.call(async (token) => {
    const { client, callOptions } = await createGrpcClient(token);
    const request = new SomeNewRequest({
      // map parameters to request object
    });
    return client.someNewEndpoint(request, callOptions);
  });
};
```

### Implementing CRUD Operations

#### Read Operations

```typescript
// Get single resource
export const getResource = async (resourceId: string) => {
  const { GetResourceRequest } = await import("./pb/resource_pb");

  return GrpcService.call(async (token) => {
    const { client, callOptions } = await createGrpcClient(token);
    return client.getResource(
      new GetResourceRequest({ name: resourceId }),
      callOptions
    );
  });
};

// List resources
export const listResources = async (
  parent: string,
  pageSize: number = 10,
  pageToken?: string
) => {
  const { ListResourcesRequest } = await import("./pb/resource_pb");

  return GrpcService.call(async (token) => {
    const { client, callOptions } = await createGrpcClient(token);
    const request = new ListResourcesRequest({
      parent,
      pageSize,
      pageToken,
    });
    return client.listResources(request, callOptions);
  });
};
```

#### Create Operations

```typescript
export const createResource = async (
  resource: Partial<Resource>,
  parent: string
) => {
  const { CreateResourceRequest } = await import("./pb/resource_pb");

  return GrpcService.call(async (token) => {
    const { client, callOptions } = await createGrpcClient(token);
    const request = new CreateResourceRequest({
      parent,
      resource: resource as Resource,
    });
    return client.createResource(request, callOptions);
  });
};
```

#### Update Operations

Always include an appropriate field mask for partial updates:

```typescript
export const updateResource = async (
  resource: Partial<Resource>,
  updateMask: string[] = []
) => {
  const { UpdateResourceRequest } = await import("./pb/resource_pb");
  const { FieldMask } = await import("./pb/google/protobuf/field_mask_pb");

  return GrpcService.call(async (token) => {
    const { client, callOptions } = await createGrpcClient(token);
    const request = new UpdateResourceRequest({
      resource: resource as Resource,
      updateMask: new FieldMask({ paths: updateMask }),
    });
    return client.updateResource(request, callOptions);
  });
};
```

#### Delete Operations

```typescript
export const deleteResource = async (resourceId: string) => {
  const { DeleteResourceRequest } = await import("./pb/resource_pb");

  return GrpcService.call(async (token) => {
    const { client, callOptions } = await createGrpcClient(token);
    return client.deleteResource(
      new DeleteResourceRequest({ name: resourceId }),
      callOptions
    );
  });
};
```

## UI Implementation Best Practices

When consuming gRPC services in React components:

### Data Fetching Pattern

```typescript
import { useState, useEffect } from "react";
import { getResource } from "@/lib/grpc-client";

function ResourceComponent({ resourceId }) {
  const [resource, setResource] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    async function fetchResource() {
      try {
        setLoading(true);
        const result = await getResource(resourceId);
        setResource(result);
      } catch (err) {
        console.error("Error fetching resource:", err);
        setError(
          `Error: ${err instanceof Error ? err.message : "Unknown error"}`
        );
      } finally {
        setLoading(false);
      }
    }

    fetchResource();
  }, [resourceId]);

  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error}</div>;
  if (!resource) return <div>Resource not found</div>;

  return (
    <div>
      <h1>{resource.title}</h1>
      {/* Render resource details */}
    </div>
  );
}
```

### Form Submission Pattern

```typescript
import { useState } from "react";
import { updateResource } from "@/lib/grpc-client";

function EditResourceForm({ resource, onSuccess }) {
  const [formData, setFormData] = useState({
    title: resource.title,
    description: resource.description,
  });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState(null);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSubmitting(true);
    setError(null);

    try {
      // Only include changed fields in the update
      const updateFields = {};
      const updateMask = [];

      if (formData.title !== resource.title) {
        updateFields.title = formData.title;
        updateMask.push("title");
      }

      if (formData.description !== resource.description) {
        updateFields.description = formData.description;
        updateMask.push("description");
      }

      // Only make the API call if there are changes
      if (updateMask.length > 0) {
        // Include the resource name for the update
        const updatedResource = await updateResource(
          {
            name: resource.name,
            ...updateFields,
          },
          updateMask
        );

        onSuccess(updatedResource);
      } else {
        onSuccess(resource); // No changes, just return the original
      }
    } catch (err) {
      console.error("Error updating resource:", err);
      setError(
        `Error: ${err instanceof Error ? err.message : "Unknown error"}`
      );
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      {error && <div className="error">{error}</div>}

      <div>
        <label htmlFor="title">Title:</label>
        <input
          id="title"
          name="title"
          value={formData.title}
          onChange={handleChange}
          disabled={submitting}
        />
      </div>

      <div>
        <label htmlFor="description">Description:</label>
        <textarea
          id="description"
          name="description"
          value={formData.description}
          onChange={handleChange}
          disabled={submitting}
        />
      </div>

      <button type="submit" disabled={submitting}>
        {submitting ? "Saving..." : "Save Changes"}
      </button>
    </form>
  );
}
```

## Best Practices

1. **Keep imports dynamic** for better code-splitting
2. **Use typed interfaces** for all request/response handling
3. **Follow resource patterns** from Google API Design Guidelines
4. **Add error handling** specific to your domain when needed
5. **Document new endpoints** as you add them
6. **Use consistent naming** across all endpoints
7. **Handle pagination properly** in list operations
8. **Use field masks correctly** for partial updates
9. **Respect resource hierarchies** (parent-child relationships)
10. **Auto-generate field masks** when possible for better developer experience
11. **Include proper error handling** in UI components
12. **Follow AIP principles** for resource naming and method signatures

## Field Mask Best Practices

When working with field masks for partial updates:

1. **Map client-side fields to protobuf field names** correctly:

   ```typescript
   // Example mapping
   if (formData.firstName !== userData.firstName) {
     updateFields.firstName = formData.firstName;
     updateMask.push("first_name"); // Use snake_case for protobuf field names
   }
   ```

2. **Auto-generate field masks from changed fields**:

   ```typescript
   // If updateMask is empty, automatically generate it from userData keys
   if (updateMask.length === 0 && userData) {
     updateMask = Object.keys(userData).filter((key) => key !== "name");
   }
   ```

3. **Always include the resource identifier** in update requests:

   ```typescript
   const request = new UpdateResourceRequest({
     resource: new Resource({
       name: resourceId, // Always include the resource name
       ...updateFields,
     }),
     updateMask: new FieldMask({ paths: updateMask }),
   });
   ```

4. **Exclude identifier fields** from update masks:
   ```typescript
   // Filter out identifier fields that shouldn't be updated
   updateMask = Object.keys(userData).filter((key) => key !== "name");
   ```

## UI Component Patterns

When building forms that interact with gRPC services:

### State Management

```typescript
// Recommended state pattern for form components
const [formData, setFormData] = useState({
  // Initialize with existing resource data
  field1: resource.field1,
  field2: resource.field2,
});
const [submitting, setSubmitting] = useState(false);
const [error, setError] = useState<string | null>(null);
const [success, setSuccess] = useState(false);
```

### Change Tracking

```typescript
// Track only changed fields for more efficient updates
const handleSubmit = async (e: React.FormEvent) => {
  e.preventDefault();
  setSubmitting(true);

  try {
    const updateFields = {};
    const updateMask = [];

    // Only include fields that have changed
    if (formData.field1 !== resource.field1) {
      updateFields.field1 = formData.field1;
      updateMask.push("field_1");
    }

    // Only make API call if there are changes
    if (updateMask.length > 0) {
      const updatedResource = await updateResource(
        {
          name: resource.name, // Always include the resource name
          ...updateFields,
        },
        updateMask
      );

      // Update local state with the returned resource
      onUpdate(updatedResource);
      setSuccess(true);
    }
  } catch (err) {
    setError(`Error: ${err instanceof Error ? err.message : "Unknown error"}`);
  } finally {
    setSubmitting(false);
  }
};
```

By following these guidelines, you'll maintain a consistent, maintainable, and robust frontend-to-backend communication pattern throughout the application.
