import React from 'react';
import Breadcrumb from '../../components/Breadcrumbs/Breadcrumb';
import ProgressFour from '../../components/Progress/ProgressFour';
import ProgressOne from '../../components/Progress/ProgressOne';
import ProgressThree from '../../components/Progress/ProgressThree';
import ProgressTwo from '../../components/Progress/ProgressTwo';

const Progress: React.FC = () => {
  return (
    <>
      <Breadcrumb pageName="Progress" />

      <div className="flex flex-col gap-7.5">
        <ProgressOne />
        <ProgressTwo />
        <ProgressThree />
        <ProgressFour />
      </div>
    </>
  );
};

export default Progress;
