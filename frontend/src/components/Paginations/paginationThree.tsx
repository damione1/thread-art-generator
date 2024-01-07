import React from 'react';
import { Link } from 'react-router-dom'

const PaginationThree: React.FC = () => {
  return (
    <div className="rounded-sm border border-stroke bg-white shadow-default dark:border-strokedark dark:bg-boxdark">
      <div className="border-b border-stroke px-4 py-4 dark:border-strokedark sm:px-6 xl:px-7.5">
        <h3 className="font-medium text-black dark:text-white">
          Pagination Style 3
        </h3>
      </div>

      <div className="p-4 sm:p-6 xl:p-7.5">
        <nav>
          <ul className="flex flex-wrap items-center gap-2">
            <li>
              <Link
                className="flex items-center justify-center rounded bg-[#EDEFF1] py-1.5 px-3 text-xs font-medium text-black hover:bg-primary hover:text-white dark:bg-graydark dark:text-white dark:hover:bg-primary dark:hover:text-white"
                to="#"
              >
                Previous
              </Link>
            </li>
            <li>
              <Link
                className="flex items-center justify-center rounded py-1.5 px-3 font-medium hover:bg-primary hover:text-white"
                to="#"
              >
                1
              </Link>
            </li>
            <li>
              <Link
                className="flex items-center justify-center rounded py-1.5 px-3 font-medium hover:bg-primary hover:text-white"
                to="#"
              >
                2
              </Link>
            </li>
            <li>
              <Link
                className="flex items-center justify-center rounded py-1.5 px-3 font-medium hover:bg-primary hover:text-white"
                to="#"
              >
                3
              </Link>
            </li>
            <li>
              <Link
                className="flex items-center justify-center rounded py-1.5 px-3 font-medium hover:bg-primary hover:text-white"
                to="#"
              >
                4
              </Link>
            </li>
            <li>
              <Link
                className="flex items-center justify-center rounded py-1.5 px-3 font-medium hover:bg-primary hover:text-white"
                to="#"
              >
                5
              </Link>
            </li>
            <li>
              <Link
                className="flex items-center justify-center rounded bg-[#EDEFF1] py-1.5 px-3 text-xs font-medium text-black hover:bg-primary hover:text-white dark:bg-graydark dark:text-white dark:hover:bg-primary dark:hover:text-white"
                to="#"
              >
                Next
              </Link>
            </li>
          </ul>
        </nav>
      </div>
    </div>
  );
};

export default PaginationThree;
