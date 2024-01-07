import React from 'react';
import Breadcrumb from '../../components/Breadcrumbs/Breadcrumb';
import PaginationOne from '../../components/Paginations/PaginationOne';
import PaginationTwo from '../../components/Paginations/PaginationTwo';
import PaginationThree from '../../components/Paginations/paginationThree';

const Pagination: React.FC = () => {
  return (
    <>
      <Breadcrumb pageName="Pagination" />

      <div className="flex flex-col gap-7.5">
        <PaginationOne />
        <PaginationTwo />
        <PaginationThree />
      </div>
    </>
  );
};

export default Pagination;
