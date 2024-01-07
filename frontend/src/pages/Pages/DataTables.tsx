import React from 'react';
import Breadcrumb from '../../components/Breadcrumbs/Breadcrumb';
import DataTableOne from '../../components/DataTables/DataTableOne';
import DataTableTwo from '../../components/DataTables/DataTableTwo';

const DataTables: React.FC = () => {
  return (
    <>
      <Breadcrumb pageName="Data Tables" />

      <div className="flex flex-col gap-5 md:gap-7 2xl:gap-10">
        <DataTableOne />
        <DataTableTwo />
      </div>
    </>
  );
};

export default DataTables;
