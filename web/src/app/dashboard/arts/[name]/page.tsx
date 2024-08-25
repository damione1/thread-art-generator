"use client";
import Breadcrumb from "@/components/Breadcrumbs/Breadcrumb";
import { notFound } from "next/navigation";
import { Loader2 } from "lucide-react";
import { useEffect, useState } from "react";
import { ArtGeneratorServiceClient } from "../../../../../grpc/ServicesServiceClientPb";
import { useSession } from "next-auth/react";
import { Art, GetArtRequest } from "../../../../../grpc/art_pb";
import { AddEditArt } from "../add-edit-form";

export default function ArtEdit({
  params,
}: {
  params: { name: string };
}) {

  if (params === undefined || params.name === undefined) {
    notFound();
  }

  const { data: session, status } = useSession({ required: true });
  const [loading, setLoading] = useState(true);
  const [art, setArt] = useState(null as unknown as Art);

  useEffect(() => {
    const fetchArt = async () => {
      if (status === "authenticated" && session?.backendTokens?.accessToken) {
        try {
          const client = new ArtGeneratorServiceClient(process.env.NEXT_PUBLIC_GRPC_API as string);
          const metadata = {
            'Authorization': 'Bearer ' + session.backendTokens.accessToken,
          };

          const getArtRequest = new GetArtRequest();
          const resourceName = `${session.user.id}/arts/${params.name}`;
          getArtRequest.setName(resourceName);

          const response = await client.getArt(getArtRequest, metadata);
          setArt(response);
        } catch (error) {
          console.error("Error getting art", error);
          notFound();
        } finally {
          setLoading(false);
        }
      }
    };

    fetchArt();
  }, [status, session, params.name]);

  if (status === "loading") {
    return <Loader2 className="animate-spin mr-4" />;
  }

  return (
    <div>
      <Breadcrumb pageName="Edit art" />
      <div className="grid grid-cols-1 gap-9">
        <div className="flex flex-col gap-9">
          {loading ? (
            <Loader2 className="animate-spin mr-4" />
          ) : (
            <AddEditArt art={art} />
          )}
        </div>
      </div>
    </div>
  );
}
