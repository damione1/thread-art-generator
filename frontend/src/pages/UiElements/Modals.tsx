import React from 'react';
import Breadcrumb from '../../components/Breadcrumbs/Breadcrumb';
import ModalOne from '../../components/Modals/ModalOne';
import ModalThree from '../../components/Modals/ModalThree';
import ModalTwo from '../../components/Modals/ModalTwo';

const Modals: React.FC = () => {
  return (
    <>
      <Breadcrumb pageName="Modals" />

      <div className="rounded-sm border border-stroke bg-white p-10 shadow-default dark:border-strokedark dark:bg-boxdark">
        <div className="flex flex-wrap justify-center gap-5">
          <ModalOne />
          <ModalTwo />
          <ModalThree />
        </div>
      </div>
    </>
  );
};

export default Modals;
