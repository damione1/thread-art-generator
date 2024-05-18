
function parseError(error: any, setError: any) {
    const errorMessages: string[] = error.message.split(';').map((err: string) => err.trim());
    errorMessages.forEach((errMsg: string) => {
      const [field, message] = errMsg.split(':').map((part: string) => part.trim());

      if (field && message) {
        setError(field, { message }, { shouldFocus: true });
      } else {
        setError("root", { message: error.message });
      }
    });
  }

  export default parseError;
