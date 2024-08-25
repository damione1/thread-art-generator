import React from 'react';
import { CardItemProps } from '../../types/cards';
import Link from 'next/link';
import Image from 'next/image';

const CardsItemOne: React.FC<CardItemProps> = ({
  imageSrc = '',
  name,
  role,
  cardImageSrc = '',
  cardTitle,
  cardContent,
}) => {
  return (
    <div className="rounded-sm border border-stroke bg-white shadow-default dark:border-strokedark dark:bg-boxdark">
      <div className="flex items-center gap-3 py-5 px-6">
        <div className="h-10 w-10 rounded-full">
          <Image src={imageSrc} alt="User" />
        </div>
        <div>
          <h4 className="font-medium text-black dark:text-white">{name}</h4>
          <p className="text-sm">{role}</p>
        </div>
      </div>

      <Link href="#" className="block px-4">
        <Image src={cardImageSrc} alt="Cards" />
      </Link>

      <div className="p-6">
        <h4 className="mb-3 text-xl font-semibold text-black hover:text-primary dark:text-white dark:hover:text-primary">
          <Link href="#">{cardTitle}</Link>
        </h4>
        <p>{cardContent}</p>
      </div>
    </div>
  );
};

export default CardsItemOne;
