import React from 'react';
import { Link } from 'react-router-dom';
import { CardItemProps } from '../types/cards';

const CardsItemTwo: React.FC<CardItemProps> = ({
  cardImageSrc,
  cardTitle,
  cardContent,
}) => {
  return (
    <div className="rounded-sm border border-stroke bg-white shadow-default dark:border-strokedark dark:bg-boxdark">
      <Link to="#" className="block px-4 pt-4">
        <img src={cardImageSrc} alt="Cards" />
      </Link>

      <div className="p-6">
        <h4 className="mb-3 text-xl font-semibold text-black hover:text-primary dark:text-white dark:hover:text-primary">
          <Link to="#">{cardTitle}</Link>
        </h4>
        <p>{cardContent}</p>
      </div>
    </div>
  );
};

export default CardsItemTwo;
