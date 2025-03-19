import { useState } from "react";
import { updateUser } from "@/lib/grpc-client";
import { User } from "@/lib/pb/user_pb";

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
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
    // Clear success/error messages when user makes changes
    setError(null);
    setSuccess(false);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError(null);
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
      setError(
        `Error: ${err instanceof Error ? err.message : "Unknown error"}`
      );
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="bg-dark-200 rounded-lg p-6 shadow-lg">
      <h2 className="text-xl font-semibold mb-4 text-slate-100">
        Edit Profile
      </h2>

      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative mb-4">
          <span className="block sm:inline">{error}</span>
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
              className="w-full p-2 bg-dark-300 border border-dark-400 rounded text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
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
              className="w-full p-2 bg-dark-300 border border-dark-400 rounded text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
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
              className="w-full p-2 bg-dark-300 border border-dark-400 rounded text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
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
              className="w-full p-2 bg-dark-300 border border-dark-400 rounded text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
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
