"use client";

import Breadcrumb from "@/components/Breadcrumbs/Breadcrumb";
import { Loader2 } from "lucide-react";
import { useSession } from "next-auth/react";
import { ArtGeneratorServiceClient } from "../../../../grpc/ServicesServiceClientPb";
import { GetUserRequest } from "../../../../grpc/user_pb";
import { useEffect, useState } from "react";
import EditProfile from "./form";

export default  function EditAccount() {
  const { data: session, status } = useSession()

  const [defaultValues, setDefaultValues] = useState({
    first_name: '',
    last_name: '',
    email: '',
  } as any);

  const [loading, setLoading] = useState(true);

  const client = new ArtGeneratorServiceClient(process.env.NEXT_PUBLIC_GRPC_API as string);
  const metadata = {
    'Authorization': 'Bearer ' + session?.backendTokens.accessToken
  }

  useEffect(() => {
    const getUserRequest = new GetUserRequest();
    getUserRequest.setName(session?.user.id as unknown as string);

    client.getUser(getUserRequest, metadata)
      .then((response) => {
        setDefaultValues({
          first_name: response.getFirstName(),
          last_name: response.getLastName(),
          email: response.getEmail(),
        });

        setLoading(false);
      })
      .catch((error) => {
        console.error("Error getting user", error);
      });
  }, []); // Empty dependency array ensures this runs once on mount


  return (
    <div>
      <Breadcrumb pageName="Edit account" />

      <div className="grid grid-cols-1 gap-9">
        <div className="flex flex-col gap-9">
          {loading ?(
            <Loader2 className="animate-spin mr-4" />
          ) : (
            <EditProfile defaultValues={defaultValues} />
          ) }
        </div>
      </div>
    </div>
  );
}
