import React from 'react';
import BadgeFour from '../../components/Badges/BadgeFour';
import BadgeOne from '../../components/Badges/BadgeOne';
import BadgeThree from '../../components/Badges/BadgeThree';
import BadgeTwo from '../../components/Badges/BadgeTwo';
import Breadcrumb from '../../components/Breadcrumbs/Breadcrumb';

const Badge: React.FC = () => {
  return (
    <>
      <Breadcrumb pageName="Badge" />

      <div className="flex flex-col gap-7.5">
        <BadgeOne />
        <BadgeTwo />
        <BadgeThree />
        <BadgeFour />
      </div>
    </>
  );
};

export default Badge;
