import Breadcrumb from "@/components/Breadcrumbs/Breadcrumb";
import { AddEditArt } from "../add-edit-form";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Create a new art",
  description: "Create a new art",
};

export default async function AdminArtCreate() {
  return (
    <div>
      <Breadcrumb pageName="Create a new art" />

      <div className="grid grid-cols-1 gap-9">
        <div className="flex flex-col gap-9">
          <AddEditArt art={null} />
        </div>
      </div>
    </div>
  );
}
