import Link from "next/link";
import { User } from "../../types/user";
import Navigation from "./Navigation";
import AuthLinks from "./AuthLinks";
import UserMenu from "./UserMenu";

interface HeaderProps {
  user?: User | null;
}

export default function Header({ user }: HeaderProps) {
  return (
    <header className="sticky top-0 z-50 border-b border-dark-300 bg-dark-100/80 backdrop-blur-md">
      <div className="container mx-auto flex items-center justify-between p-4">
        <div className="flex items-center gap-3">
          <Link
            href="/"
            className="text-2xl font-bold tracking-tight text-slate-100"
          >
            Thread<span className="text-primary-500">Art</span>
          </Link>
        </div>

        <Navigation user={user} />

        <div className="flex items-center space-x-4">
          {!user ? <AuthLinks /> : <UserMenu user={user} />}
        </div>
      </div>
    </header>
  );
}
