export default function Loading() {
    return (
        <div className="flex h-screen">
            <div className="m-auto">
                <div className="flex justify-center">
                    <div className="animate-spin rounded-full h-32 w-32 border-t-2 border-b-2 border-gray-900"></div>
                </div>
                <div className="text-center mt-4 text-gray-600">Loading...</div>
            </div>
        </div>
    )
}
