import React from 'react';
import CoverThree from '../../images/cover/cover-03.jpg';
import CoverFour from '../../images/cover/cover-04.jpg';
import CoverFive from '../../images/cover/cover-05.jpg'

const ImagesTwo: React.FC = () => {
  return (
    <div className="rounded-sm border border-stroke bg-white shadow-default dark:border-strokedark dark:bg-boxdark">
      <div className="border-b border-stroke px-4 py-4 dark:border-strokedark sm:px-6 xl:px-7.5">
        <h3 className="font-medium text-black dark:text-white">Image Grid</h3>
      </div>

      <div className="p-4 sm:p-6 xl:p-10">
        <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-3 xl:gap-10">
          <div>
            <img src={CoverThree} alt="Cover" />
          </div>
          <div>
            <img src={CoverFour} alt="Cover" />
          </div>
          <div>
            <img src={CoverFive} alt="Cover" />
          </div>
        </div>
      </div>
    </div>
  );
};

export default ImagesTwo;
