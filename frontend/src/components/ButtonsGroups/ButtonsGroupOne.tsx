import React from 'react';
import { Link } from 'react-router-dom';

const ButtonsGroupOne: React.FC = () => {
  return (
    <div className="rounded-sm border border-stroke bg-white shadow-default dark:border-strokedark dark:bg-boxdark">
      <div className="border-b border-stroke px-4 py-4 dark:border-strokedark sm:px-6 xl:px-7.5">
        <h3 className="font-medium text-black dark:text-white">
          Buttons Group
        </h3>
      </div>

      <div className="p-4 sm:p-5 xl:p-7.5">
        <div className="flex flex-col gap-6">
          <div className="flex items-center">
            <Link
              to="#"
              className="inline-flex border border-primary bg-primary py-1 px-2 font-medium text-white hover:border-primary hover:bg-primary hover:text-white dark:hover:border-primary sm:py-3 sm:px-6"
            >
              About
            </Link>
            <Link
              to="#"
              className="inline-flex border-y border-stroke py-1 px-2 font-medium text-black hover:border-primary hover:bg-primary hover:text-white dark:border-strokedark dark:text-white dark:hover:border-primary sm:py-3 sm:px-6"
            >
              Profile
            </Link>
            <Link
              to="#"
              className="inline-flex border border-stroke py-1 px-2 font-medium text-black hover:border-primary hover:bg-primary hover:text-white dark:border-strokedark dark:text-white dark:hover:border-primary sm:py-3 sm:px-6"
            >
              Services
            </Link>
          </div>

          <div className="flex items-center rounded-lg">
            <Link
              to="#"
              className="inline-flex rounded-l-lg border border-primary bg-primary py-1 px-2 font-medium text-white hover:border-primary hover:bg-primary hover:text-white dark:hover:border-primary sm:py-3 sm:px-6"
            >
              About
            </Link>
            <Link
              to="#"
              className="inline-flex border-y border-stroke py-1 px-2 font-medium text-black hover:border-primary hover:bg-primary hover:text-white dark:border-strokedark dark:text-white dark:hover:border-primary sm:py-3 sm:px-6"
            >
              Profile
            </Link>
            <Link
              to="#"
              className="inline-flex rounded-r-lg border border-stroke py-1 px-2 font-medium text-black hover:border-primary hover:bg-primary hover:text-white dark:border-strokedark dark:text-white dark:hover:border-primary sm:py-3 sm:px-6"
            >
              Services
            </Link>
          </div>

          <div className="flex items-center rounded-full">
            <Link
              to="#"
              className="inline-flex rounded-l-full border border-primary bg-primary py-1 px-2 font-medium text-white hover:border-primary hover:bg-primary hover:text-white dark:hover:border-primary sm:py-3 sm:px-6"
            >
              About
            </Link>
            <Link
              to="#"
              className="inline-flex border-y border-stroke py-1 px-2 font-medium text-black hover:border-primary hover:bg-primary hover:text-white dark:border-strokedark dark:text-white dark:hover:border-primary sm:py-3 sm:px-6"
            >
              Profile
            </Link>
            <Link
              to="#"
              className="inline-flex rounded-r-full border border-stroke py-1 px-2 font-medium text-black hover:border-primary hover:bg-primary hover:text-white dark:border-strokedark dark:text-white dark:hover:border-primary sm:py-3 sm:px-6"
            >
              Services
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ButtonsGroupOne;
