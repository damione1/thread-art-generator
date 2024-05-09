// import { ArtGeneratorServiceClient } from "@/grpc/ServicesServiceClientPb";
// import React, { useState } from "react";

// const GrpcContext = React.createContext<{
//   client: ArtGeneratorServiceClient;
// } | null>(null);

// export function GrpcProvider({ children }: { children: React.ReactNode }) {
//   const client = new ArtGeneratorServiceClient("http://localhost:9091");

//   return (
//     <GrpcContext.Provider value={{ client }}>{children}</GrpcContext.Provider>
//   );
// }

// export function useGrpcClient(): { client: ArtGeneratorServiceClient } {
//   const context = React.useContext(GrpcContext);
//   if (context === null) {
//     throw new Error("useGrpcClient must be used within a GrpcProvider");
//   }
//   return context;
// }
