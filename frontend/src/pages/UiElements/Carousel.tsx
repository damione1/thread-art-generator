import React from 'react';
import Breadcrumb from '../../components/Breadcrumbs/Breadcrumb';
import CarouselOne from '../../components/Carousels/CarouselOne';
import CarouselThree from '../../components/Carousels/CarouselThree';
import CarouselTwo from '../../components/Carousels/CarouselTwo';

const Carousel: React.FC = () => {
  return (
    <>
      <Breadcrumb pageName="Carousel" />

      <div className="flex flex-col gap-7.5">
        <CarouselOne />
        <CarouselTwo />
        <CarouselThree />
      </div>
    </>
  );
};

export default Carousel;
