export const formatDate = (dateString: string, complete: boolean = false) => {
  const date = new Date(dateString);
  let options = {
    day: "numeric",
    month: "short",
    year: "numeric",
  } as Intl.DateTimeFormatOptions;
  if (complete) {
    options = {
      day: "numeric",
      month: "short",
      year: "numeric",
      hour: "numeric",
      minute: "numeric",
      second: "numeric",
    };
  }
  return new Intl.DateTimeFormat("en", options).format(date);
};
