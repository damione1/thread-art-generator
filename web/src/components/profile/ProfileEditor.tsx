import { useState } from "react";
import { updateUser } from "@/lib/grpc-client";
import { User } from "@/lib/pb/user_pb";
import { parseErrors } from "@/utils/errorUtils";

interface ProfileEditorProps {
  userData: User;
  onUpdate: (updatedUser: User) => void;
}

export default function ProfileEditor({
  userData,
  onUpdate,
}: ProfileEditorProps) {
  const [formData, setFormData] = useState({
    firstName: userData.firstName,
    lastName: userData.lastName,
    email: userData.email,
    avatar: userData.avatar,
  });
  const [submitting, setSubmitting] = useState(false);
  const [generalError, setGeneralError] = useState<string | null>(null);
  const [fieldErrors, setFieldErrors] = useState<{ [key: string]: string }>({});
  const [success, setSuccess] = useState(false);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
    // Clear error for this field when user makes changes
    if (fieldErrors[name]) {
      setFieldErrors((prev) => {
        const newErrors = { ...prev };
        delete newErrors[name];
        return newErrors;
      });
    }
    // Clear general messages
    setGeneralError(null);
    setSuccess(false);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setGeneralError(null);
    setFieldErrors({});
    setSuccess(false);

    try {
      // Send all fields in the update
      const updatedUser = await updateUser({
        name: userData.name,
        firstName: formData.firstName || "",
        lastName: formData.lastName || "",
        email: formData.email || "",
        avatar: formData.avatar || "",
      });

      onUpdate(updatedUser);
      setSuccess(true);
    } catch (err) {
      console.error("Error updating profile:", err);

      // Parse validation errors from the API
      const errorMessage = err instanceof Error ? err.message : "Unknown error";
      const parsedErrors = parseErrors(errorMessage);

      // If we have field-specific errors, show them
      if (Object.keys(parsedErrors).length > 0) {
        // Check for general error message
        if (parsedErrors._general) {
          setGeneralError(parsedErrors._general);
          delete parsedErrors._general;
        }

        // Set remaining field errors
        setFieldErrors(parsedErrors);
      } else {
        // Fallback to general error message if no field errors are found
        setGeneralError(`Error: ${errorMessage}`);
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="bg-dark-200 rounded-lg p-6 shadow-lg">
      <h2 className="text-xl font-semibold mb-4 text-slate-100">
        Edit Profile
      </h2>

      {generalError && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative mb-4">
          <span className="block sm:inline">{generalError}</span>
        </div>
      )}

      {success && (
        <div className="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded relative mb-4">
          <span className="block sm:inline">Profile updated successfully!</span>
        </div>
      )}

      <form onSubmit={handleSubmit}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
          <div>
            <label
              htmlFor="firstName"
              className="block text-sm font-medium text-slate-400 mb-1"
            >
              First Name
            </label>
            <input
              type="text"
              id="firstName"
              name="firstName"
              value={formData.firstName}
              onChange={handleChange}
              disabled={submitting}
              className={`w-full p-2 bg-dark-300 border ${
                fieldErrors.firstName ? "border-red-500" : "border-dark-400"
              } rounded text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500`}
            />
            {fieldErrors.firstName && (
              <p className="text-red-500 text-xs mt-1">
                {fieldErrors.firstName}
              </p>
            )}
          </div>

          <div>
            <label
              htmlFor="lastName"
              className="block text-sm font-medium text-slate-400 mb-1"
            >
              Last Name
            </label>
            <input
              type="text"
              id="lastName"
              name="lastName"
              value={formData.lastName}
              onChange={handleChange}
              disabled={submitting}
              className={`w-full p-2 bg-dark-300 border ${
                fieldErrors.lastName ? "border-red-500" : "border-dark-400"
              } rounded text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500`}
            />
            {fieldErrors.lastName && (
              <p className="text-red-500 text-xs mt-1">
                {fieldErrors.lastName}
              </p>
            )}
          </div>

          <div>
            <label
              htmlFor="email"
              className="block text-sm font-medium text-slate-400 mb-1"
            >
              Email
            </label>
            <input
              type="email"
              id="email"
              name="email"
              value={formData.email}
              onChange={handleChange}
              disabled={submitting}
              className={`w-full p-2 bg-dark-300 border ${
                fieldErrors.email ? "border-red-500" : "border-dark-400"
              } rounded text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500`}
            />
            {fieldErrors.email && (
              <p className="text-red-500 text-xs mt-1">{fieldErrors.email}</p>
            )}
          </div>

          <div>
            <label
              htmlFor="avatar"
              className="block text-sm font-medium text-slate-400 mb-1"
            >
              Avatar URL
            </label>
            <input
              type="url"
              id="avatar"
              name="avatar"
              value={formData.avatar}
              onChange={handleChange}
              disabled={submitting}
              className={`w-full p-2 bg-dark-300 border ${
                fieldErrors.avatar ? "border-red-500" : "border-dark-400"
              } rounded text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500`}
            />
            {fieldErrors.avatar && (
              <p className="text-red-500 text-xs mt-1">{fieldErrors.avatar}</p>
            )}
          </div>
        </div>

        <div className="flex justify-end">
          <button
            type="submit"
            disabled={submitting}
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
          >
            {submitting ? "Saving..." : "Save Changes"}
          </button>
        </div>
      </form>
    </div>
  );
}
