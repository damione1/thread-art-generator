import Link from "next/link";

export default function Footer() {
  return (
    <footer className="w-full border-t border-dark-300 bg-dark-200 mt-auto">
      <div className="container mx-auto p-6">
        <div className="flex flex-col md:flex-row justify-between items-center">
          <div className="mb-4 md:mb-0">
            <Link
              href="/"
              className="text-xl font-bold tracking-tight text-slate-100"
            >
              Thread<span className="text-primary-500">Art</span>
            </Link>
            <p className="mt-2 text-sm text-slate-400">
              Generate beautiful thread art from images
            </p>
          </div>
          <div className="flex space-x-6">
            <a href="#" className="text-slate-400 hover:text-white transition">
              Privacy
            </a>
            <a href="#" className="text-slate-400 hover:text-white transition">
              Terms
            </a>
            <a href="#" className="text-slate-400 hover:text-white transition">
              Contact
            </a>
          </div>
        </div>
        <div className="mt-8 border-t border-dark-300 pt-8 text-center text-sm text-slate-400">
          © {new Date().getFullYear()} ThreadArt. All rights reserved.
        </div>
      </div>
    </footer>
  );
}
