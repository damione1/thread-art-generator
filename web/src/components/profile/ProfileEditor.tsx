import { updateUser } from "@/lib/grpc-client";
import { User } from "@/lib/pb/user_pb";
import {
  FormField,
  TextInput,
  EmailInput,
  ErrorMessage,
  SuccessMessage,
  useForm,
} from "@/components/ui";

export interface ProfileEditorProps {
  userData?: User;
  onUpdate?: (updatedUser: User) => void;
  onCancel?: () => void;
  onSuccess?: () => void;
}

interface ProfileFormValues extends Record<string, unknown> {
  firstName: string;
  lastName: string;
  email: string;
}

export default function ProfileEditor({
  userData,
  onUpdate,
  onCancel,
  onSuccess,
}: ProfileEditorProps) {
  // Initialize form with useForm hook
  const {
    values,
    errors,
    handleChange,
    handleSubmit,
    isSubmitting,
    generalError,
    isSuccess,
  } = useForm<ProfileFormValues>({
    firstName: userData?.firstName || "",
    lastName: userData?.lastName || "",
    email: userData?.email || "",
  });

  // Handle form submission
  const onSubmitForm = async (formValues: ProfileFormValues) => {
    try {
      if (!userData?.name) {
        throw new Error("User data is required");
      }

      const updatedUser = await updateUser({
        name: userData.name,
        firstName: formValues.firstName || "",
        lastName: formValues.lastName || "",
        email: formValues.email || "",
      });

      // Call the provided callbacks
      if (onUpdate) {
        onUpdate(updatedUser);
      }

      if (onSuccess) {
        onSuccess();
      }
    } catch (error) {
      // Error handling is now managed by the useForm hook
      throw error;
    }
  };

  return (
    <div className="bg-dark-200 rounded-lg p-6 shadow-lg">
      <h2 className="text-xl font-semibold mb-4 text-slate-100">
        Edit Profile
      </h2>

      <ErrorMessage message={generalError} />
      <SuccessMessage
        message={isSuccess ? "Profile updated successfully!" : null}
      />

      <form onSubmit={handleSubmit(onSubmitForm)}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
          <FormField
            id="firstName"
            name="firstName"
            label="First Name"
            error={errors.firstName}
            inputComponent={
              <TextInput
                id="firstName"
                name="firstName"
                value={values.firstName}
                onChange={handleChange}
                disabled={isSubmitting}
              />
            }
          />

          <FormField
            id="lastName"
            name="lastName"
            label="Last Name"
            error={errors.lastName}
            inputComponent={
              <TextInput
                id="lastName"
                name="lastName"
                value={values.lastName}
                onChange={handleChange}
                disabled={isSubmitting}
              />
            }
          />

          <FormField
            id="email"
            name="email"
            label="Email"
            error={errors.email}
            inputComponent={
              <EmailInput
                id="email"
                name="email"
                value={values.email}
                onChange={handleChange}
                disabled={isSubmitting}
              />
            }
          />
        </div>

        <div className="flex justify-end gap-3">
          {onCancel && (
            <button
              type="button"
              onClick={onCancel}
              disabled={isSubmitting}
              className="px-4 py-2 bg-dark-300 text-white rounded hover:bg-dark-400 transition"
            >
              Cancel
            </button>
          )}
          <button
            type="submit"
            disabled={isSubmitting}
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
          >
            {isSubmitting ? "Saving..." : "Save Changes"}
          </button>
        </div>
      </form>
    </div>
  );
}
