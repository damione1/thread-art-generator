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
    firstName: '',
    lastName: '',
    email: '',
  } as any);

  const [loading, setLoading] = useState(true);

  const client = new ArtGeneratorServiceClient("http://localhost:8080");
  const metadata = {
    'Authorization': 'Bearer ' + session?.backendTokens.accessToken
  }

  useEffect(() => {
    const getUserRequest = new GetUserRequest();
    getUserRequest.setName(session?.user.id as unknown as string);

    client.getUser(getUserRequest, metadata)
      .then((response) => {
        setDefaultValues({
          firstName: response.getFirstName(),
          lastName: response.getLastName(),
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
      <Breadcrumb pageName="Edit Project" />

      <div className="grid grid-cols-1 gap-9">
        <div className="flex flex-col gap-9">
          {loading ?(
            <span className="animate-spin mr-4">
              <Loader2 />
            </span>
          ) : (
            <EditProfile defaultValues={defaultValues} />
          ) }
        </div>
      </div>
    </div>
  );
}
