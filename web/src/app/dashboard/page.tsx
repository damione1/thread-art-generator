"use client";
import Breadcrumb from "@/components/Breadcrumbs/Breadcrumb";
import CardsItemTwo from "@/components/Cards/CardsItemTwo";
import { Metadata } from "next";
import { useSession } from "next-auth/react";
import { ArtGeneratorServiceClient } from "../../../grpc/ServicesServiceClientPb";
import { Art, GetArtRequest, ListArtsRequest } from "../../../grpc/art_pb";
import { useEffect, useState } from "react";
import { parseResourceName } from "../../../util/ressourceName";


export default function Home() {
  const { data: session, status } = useSession({ required: true })
  const client = new ArtGeneratorServiceClient(process.env.NEXT_PUBLIC_GRPC_API as string);
  const metadata = {
    'Authorization': 'Bearer ' + session?.backendTokens.accessToken
  }

  const [artsList, setArtsList ] = useState([] as Array<Art>);

  useEffect(() => {
    const listArtsRequest = new ListArtsRequest();
    listArtsRequest.setParent(session?.user.id as unknown as string);
    listArtsRequest.setPageSize(10);
    listArtsRequest.setPageToken(0);

    client.listArts(listArtsRequest, metadata)
      .then((response) => {
        setArtsList(response.getArtsList());
      })
      .catch((error) => {
        console.error("Error getting arts list", error);
      });
  }, []); // Empty dependency array ensures this runs once on mount


  return (
    <>
    <Breadcrumb pageName="Home" />

  <div className="grid grid-cols-1 gap-7.5 sm:grid-cols-2 xl:grid-cols-3">
    <CardsItemTwo
        cardImageSrc=""
        cardTitle= "Create new art"
        cardContent=""
        cardLink="/dashboard/arts/new"
      />
    {artsList.map((art) => (
      <>
      <CardsItemTwo
        key={art.getName()}
        cardImageSrc={art.getImageUrl()}
        cardTitle={art.getTitle()}
        cardContent={art.getCreateTime().toDate().toLocaleDateString()}
        cardLink={`/dashboard/arts/` + parseResourceName(art.getName()).get("arts")}
      />
      </>
    ))}
  </div>
  </>
  );
}
