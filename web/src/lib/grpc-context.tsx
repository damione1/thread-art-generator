"use client";
import React, { useState } from "react";
import { ArtGeneratorServiceClient } from "../../grpc/ServicesServiceClientPb";

const GrpcContext = React.createContext<{
  client: ArtGeneratorServiceClient;
} | null>(null);

export function GrpcProvider({ children }: { children: React.ReactNode }) {

  const client = new ArtGeneratorServiceClient("http://localhost:8080");

  return (
    <GrpcContext.Provider value={{ client }}>{children}</GrpcContext.Provider>
  );
}

export function UseGrpcClient(): { client: ArtGeneratorServiceClient } {
  const context = React.useContext(GrpcContext);
  if (context === null) {
    throw new Error("useGrpcClient must be used within a GrpcProvider");
  }
  return context;
}
