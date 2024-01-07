import React from 'react';
import Banner from '../../images/best-value-banner.png'

const PricingTableOne: React.FC = () => {
  return (
    <div className="grid grid-cols-1 gap-4 md:grid-cols-2 md:gap-6 xl:grid-cols-3 2xl:gap-7.5">
      {/* <!-- Pricing Item --> */}
      <div className="relative rounded-sm border border-stroke bg-white p-6 shadow-default dark:border-strokedark dark:bg-boxdark md:p-9 xl:p-11.5">
        <span className="mb-2.5 block text-title-sm2 font-bold text-black dark:text-white">
          Starter
        </span>
        <h3>
          <span className="text-xl font-medium text-black dark:text-white">
            $
          </span>
          <span className="text-title-xxl2 font-bold text-black dark:text-white">
            25.00
          </span>
          <span className="font-medium"> Per Month</span>
        </h3>

        <h4 className="mt-7.5 mb-5 text-lg font-medium text-black dark:text-white">
          Features
        </h4>

        <ul className="flex flex-col gap-3.5">
          <li className="font-medium">Get Figma Source File</li>
          <li className="font-medium">Use on Unlimited Projects</li>
          <li className="font-medium">Personal & Commercial Use</li>
          <li className="font-medium">Downloadable Offline Files</li>
          <li className="font-medium">Lifetime Free Updates</li>
        </ul>

        <button className="mt-9 flex rounded-md border border-primary py-3 px-9 font-medium text-primary hover:bg-primary hover:text-white">
          Purchase Now
        </button>
      </div>

      {/* <!-- Pricing Item --> */}
      <div className="relative rounded-sm border border-stroke bg-white p-6 shadow-default dark:border-strokedark dark:bg-boxdark md:p-9 xl:p-11.5">
        <span className="absolute top-3.5 -right-1">
          <img src={Banner} alt="Banner" />
        </span>
        <span className="mb-2.5 block text-title-sm2 font-bold text-black dark:text-white">
          Basic
        </span>
        <h3>
          <span className="text-xl font-medium text-black dark:text-white">
            $
          </span>
          <span className="text-title-xxl2 font-bold text-black dark:text-white">
            59.00
          </span>
          <span className="font-medium"> Per Month</span>
        </h3>

        <h4 className="mt-7.5 mb-5 text-lg font-medium text-black dark:text-white">
          Features
        </h4>

        <ul className="flex flex-col gap-3.5">
          <li className="font-medium">Get Figma Source File</li>
          <li className="font-medium">Use on Unlimited Projects</li>
          <li className="font-medium">Personal & Commercial Use</li>
          <li className="font-medium">Downloadable Offline Files</li>
          <li className="font-medium">Lifetime Free Updates</li>
        </ul>

        <button className="mt-9 flex rounded-md border border-primary py-3 px-9 font-medium text-primary hover:bg-primary hover:text-white">
          Purchase Now
        </button>
      </div>

      {/* <!-- Pricing Item --> */}
      <div className="relative rounded-sm border border-stroke bg-white p-6 shadow-default dark:border-strokedark dark:bg-boxdark md:p-9 xl:p-11.5">
        <span className="mb-2.5 block text-title-sm2 font-bold text-black dark:text-white">
          Premium
        </span>
        <h3>
          <span className="text-xl font-medium text-black dark:text-white">
            $
          </span>
          <span className="text-title-xxl2 font-bold text-black dark:text-white">
            99.00
          </span>
          <span className="font-medium"> Per Month</span>
        </h3>

        <h4 className="mt-7.5 mb-5 text-lg font-medium text-black dark:text-white">
          Features
        </h4>

        <ul className="flex flex-col gap-3.5">
          <li className="font-medium">Get Figma Source File</li>
          <li className="font-medium">Use on Unlimited Projects</li>
          <li className="font-medium">Personal & Commercial Use</li>
          <li className="font-medium">Downloadable Offline Files</li>
          <li className="font-medium">Lifetime Free Updates</li>
        </ul>

        <button className="mt-9 flex rounded-md border border-primary py-3 px-9 font-medium text-primary hover:bg-primary hover:text-white">
          Purchase Now
        </button>
      </div>
    </div>
  );
};

export default PricingTableOne;
