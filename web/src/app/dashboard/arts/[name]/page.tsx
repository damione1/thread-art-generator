"use client";
import Breadcrumb from "@/components/Breadcrumbs/Breadcrumb";
import { notFound } from "next/navigation";
import { Metadata } from "next";
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

  const { data: session, status } = useSession({ required: true })
  const [loading, setLoading] = useState(true);
  const [art, setArt] = useState(null as unknown as Art);
  const client = new ArtGeneratorServiceClient(process.env.NEXT_PUBLIC_GRPC_API as string);
  const metadata = {
    'Authorization': 'Bearer ' + session?.backendTokens.accessToken
  }

  useEffect(() => {
    const getArtRequest = new GetArtRequest();
    const ressourceName = `${session?.user.id}/arts/${params.name}`;

    getArtRequest.setName(ressourceName);

    client.getArt(getArtRequest, metadata)
      .then((response) => {
        setArt(response);
        setLoading(false);
      })
      .catch((error) => {
          console.error("Error getting art", error);
          notFound();
      });
  }, []); // Empty dependency array ensures this runs once on mount

  return (
    <div>
      <Breadcrumb pageName="Edit art" />

      <div className="grid grid-cols-1 gap-9">
        <div className="flex flex-col gap-9">
          {art ? (
            <AddEditArt art={art} />
          ) : (
              <Loader2 className="animate-spin mr-4" />
          )}
        </div>
      </div>
    </div>
  );
}
