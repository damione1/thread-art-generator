import { CardItemProps } from '@/types/cards';
import Link from 'next/link';
import React from 'react';

const CardsItemThree: React.FC<CardItemProps> = ({
  cardTitle,
  cardContent,
}) => {
  return (
    <div className="rounded-sm border border-stroke bg-white shadow-default dark:border-strokedark dark:bg-boxdark">
      <div className="border-b border-stroke p-5 px-7.5 dark:border-strokedark">
        <h4 className="text-xl font-semibold text-black hover:text-primary dark:text-white dark:hover:text-primary">
          <Link href="#">{cardTitle}</Link>
        </h4>
      </div>
      <div className="px-7.5 pt-6 pb-9">
        <p>{cardContent}</p>
      </div>
    </div>
  );
};

export default CardsItemThree;
