function parseResourceName(resourceName: string): Map<string, string> {
    const resourceMap = new Map<string, string>();
    const segments = resourceName.split('/');

    if (segments.length % 2 !== 0) {
        throw new Error('Invalid resource name format');
    }

    for (let i = 0; i < segments.length; i += 2) {
        const resourceType = segments[i];
        const resourceId = segments[i + 1];

        if (!isValidUUID(resourceId)) {
            throw new Error(`Invalid UUID format for resource id: ${resourceId}`);
        }

        resourceMap.set(resourceType, resourceId);
    }

    console.log(resourceMap);
    return resourceMap;
}


function isValidUUID(uuid: string): boolean {
    const uuidRegex = /^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$/;
    return uuidRegex.test(uuid);
}



export { parseResourceName };
