/**
 * Get the display information for an art status
 * @param status The status code of the art
 * @returns Object containing text and color for the status
 */
export function getStatusInfo(status: number) {
    switch (status) {
        case 3: // ART_STATUS_COMPLETE
            return { text: "Completed", color: "text-accent-teal" };
        case 2: // ART_STATUS_PROCESSING
            return { text: "Processing", color: "text-amber-400" };
        case 1: // ART_STATUS_PENDING_IMAGE
            return { text: "Pending Image", color: "text-primary-400" };
        case 4: // ART_STATUS_FAILED
            return { text: "Failed", color: "text-red-500" };
        case 5: // ART_STATUS_ARCHIVED
            return { text: "Archived", color: "text-slate-400" };
        default:
            return { text: "Unknown", color: "text-slate-400" };
    }
}
