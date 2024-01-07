import React from 'react';
import Breadcrumb from '../../components/Breadcrumbs/Breadcrumb';
import NotificationsOne from '../../components/Notifications/NotificationsOne';
import NotificationsThree from '../../components/Notifications/NotificationsThree';
import NotificationsTwo from '../../components/Notifications/NotificationsTwo';

const Notifications: React.FC = () => {
  return (
    <>
      <Breadcrumb pageName="Notifications" />

      <div className="flex flex-col gap-7.5">
        <NotificationsOne />
        <NotificationsTwo />
        <NotificationsThree />
      </div>
    </>
  );
};

export default Notifications;
