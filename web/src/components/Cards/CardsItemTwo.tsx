import React from 'react';

import Image from 'next/image';
import Link from 'next/link';
import { CardItemProps } from '@/types/cards';

const CardsItemTwo: React.FC<CardItemProps> = ({
  cardImageSrc = '',
  cardTitle,
  cardContent,
  cardLink = '#',
}) => {
  return (
    <div className="rounded-sm border border-stroke bg-white shadow-default dark:border-strokedark dark:bg-boxdark">
      <Link href={cardLink} className="block px-4 pt-4">
        <Image src={cardImageSrc} alt="Art" width={432} height={432} loading='lazy' />
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

export default CardsItemTwo;
