import { lazy, useEffect, useState } from 'react';
import { Route, Routes, useNavigate } from 'react-router-dom';

import ECommerce from './pages/Dashboard/ECommerce';
import Analytics from './pages/Dashboard/Analytics';
import Marketing from './pages/Dashboard/Marketing';
import CRM from './pages/Dashboard/CRM';
import Calendar from './pages/Calendar';
import Profile from './pages/Profile';
import TaskKanban from './pages/Task/TaskKanban';
import TaskList from './pages/Task/TaskList';
import FormElements from './pages/Form/FormElements';
import FormLayout from './pages/Form/FormLayout';
import Tables from './pages/Tables';
import Settings from './pages/Pages/Settings';
import FileManager from './pages/Pages/FileManager';
import DataTables from './pages/Pages/DataTables';
import PricingTables from './pages/Pages/PricingTables';
import ErrorPage from './pages/Pages/ErrorPage';
import MailSuccess from './pages/Pages/MailSuccess';
import Messages from './pages/Messages';
import Inbox from './pages/Inbox';
import Invoice from './pages/Invoice';
import BasicChart from './pages/Chart/BasicChart';
import AdvancedChart from './pages/Chart/AdvancedChart';
import Alerts from './pages/UiElements/Alerts';
import Buttons from './pages/UiElements/Buttons';
import ButtonsGroup from './pages/UiElements/ButtonsGroup';
import Badge from './pages/UiElements/Badge';
import Breadcrumbs from './pages/UiElements/Breadcrumbs';
import Cards from './pages/UiElements/Cards';
import Dropdowns from './pages/UiElements/Dropdowns';
import Modals from './pages/UiElements/Modals';
import Tabs from './pages/UiElements/Tabs';
import Tooltips from './pages/UiElements/Tooltips';
import Popovers from './pages/UiElements/Popovers';
import Accordion from './pages/UiElements/Accordion';
import Notifications from './pages/UiElements/Notifications';
import Pagination from './pages/UiElements/Pagination';
import Progress from './pages/UiElements/Progress';
import Carousel from './pages/UiElements/Carousel';
import Images from './pages/UiElements/Images';
import Videos from './pages/UiElements/Videos';
import ResetPassword from './pages/Authentication/ResetPassword';
import SignIn from './pages/Authentication/SignIn';
import SignUp from './pages/Authentication/SignUp';
import Loader from './common/Loader';
import PageTitle from './components/PageTitle';

const DefaultLayout = lazy(() => import('./layout/DefaultLayout'));


function App() {
  const [loading, setLoading] = useState<boolean>(true);
  const navigate = useNavigate();

  useEffect(() => {
    setTimeout(() => setLoading(false), 1000);
    const token = localStorage.getItem('loginToken');
    if (!token) navigate('/auth/signin');
  }, []);

  return loading ? (
    <Loader />
  ) : (
    <>
      <Routes>
        <Route element={<DefaultLayout />}>
          <Route
            index
            element={
              <>
                <PageTitle title="eCommerce Dashboard | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <ECommerce />
              </>
            }
          />
          <Route
            path="/dashboard/analytics"
            element={
              <>
                <PageTitle title="Analytics Dashboard | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Analytics />
              </>
            }
          />
          <Route
            path="/dashboard/marketing"
            element={
              <>
                <PageTitle title="Marketing Dashboard | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Marketing />
              </>
            }
          />
          <Route
            path="/dashboard/crm"
            element={
              <>
                <PageTitle title="CRM Dashboard | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <CRM />
              </>
            }
          />
          <Route
            path="/calendar"
            element={
              <>
                <PageTitle title="Calendar | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Calendar />
              </>
            }
          />
          <Route
            path="/profile"
            element={
              <>
                <PageTitle title="Profile | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Profile />
              </>
            }
          />
          <Route
            path="/tasks/task-list"
            element={
              <>
                <PageTitle title="Task List | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <TaskList />
              </>
            }
          />
          <Route
            path="/tasks/task-kanban"
            element={
              <>
                <PageTitle title="Task Kanban | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <TaskKanban />
              </>
            }
          />
          <Route
            path="/forms/form-elements"
            element={
              <>
                <PageTitle title="Form Elements | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <FormElements />
              </>
            }
          />
          <Route
            path="/forms/form-layout"
            element={
              <>
                <PageTitle title="Form Layout | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <FormLayout />
              </>
            }
          />
          <Route
            path="/tables"
            element={
              <>
                <PageTitle title="Tables | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Tables />
              </>
            }
          />
          <Route
            path="/pages/settings"
            element={
              <>
                <PageTitle title="Settings | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Settings />
              </>
            }
          />
          <Route
            path="/pages/file-manager"
            element={
              <>
                <PageTitle title="File Manager | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <FileManager />
              </>
            }
          />
          <Route
            path="/pages/data-tables"
            element={
              <>
                <PageTitle title="Data Tables | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <DataTables />
              </>
            }
          />
          <Route
            path="/pages/pricing-tables"
            element={
              <>
                <PageTitle title="Pricing Tables | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <PricingTables />
              </>
            }
          />
          <Route
            path="/pages/error-page"
            element={
              <>
                <PageTitle title="Error Page | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <ErrorPage />
              </>
            }
          />
          <Route
            path="/pages/mail-success"
            element={
              <>
                <PageTitle title="Mail Success | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <MailSuccess />
              </>
            }
          />
          <Route
            path="/messages"
            element={
              <>
                <PageTitle title="Messages | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Messages />
              </>
            }
          />
          <Route
            path="/inbox"
            element={
              <>
                <PageTitle title="Inbox | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Inbox />
              </>
            }
          />
          <Route
            path="/invoice"
            element={
              <>
                <PageTitle title="Invoice | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Invoice />
              </>
            }
          />
          <Route
            path="/chart/basic-chart"
            element={
              <>
                <PageTitle title="Basic Chart | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <BasicChart />
              </>
            }
          />
          <Route
            path="/chart/advanced-chart"
            element={
              <>
                <PageTitle title="Advanced Chart | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <AdvancedChart />
              </>
            }
          />
          <Route
            path="/ui/alerts"
            element={
              <>
                <PageTitle title="Alerts | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Alerts />
              </>
            }
          />
          <Route
            path="/ui/buttons"
            element={
              <>
                <PageTitle title="Buttons | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Buttons />
              </>
            }
          />
          <Route
            path="/ui/buttons-group"
            element={
              <>
                <PageTitle title="Buttons Groupo | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <ButtonsGroup />
              </>
            }
          />
          <Route
            path="/ui/badge"
            element={
              <>
                <PageTitle title="Badge | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Badge />
              </>
            }
          />
          <Route
            path="/ui/breadcrumbs"
            element={
              <>
                <PageTitle title="Breadcrumbs | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Breadcrumbs />
              </>
            }
          />
          <Route
            path="/ui/cards"
            element={
              <>
                <PageTitle title="Cards | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Cards />
              </>
            }
          />
          <Route
            path="/ui/dropdowns"
            element={
              <>
                <PageTitle title="Dropdowns | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Dropdowns />
              </>
            }
          />
          <Route
            path="/ui/modals"
            element={
              <>
                <PageTitle title="Modals | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Modals />
              </>
            }
          />
          <Route
            path="/ui/tabs"
            element={
              <>
                <PageTitle title="Tabs | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Tabs />
              </>
            }
          />
          <Route
            path="/ui/tooltips"
            element={
              <>
                <PageTitle title="Tooltips | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Tooltips />
              </>
            }
          />
          <Route
            path="/ui/popovers"
            element={
              <>
                <PageTitle title="Popovers | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Popovers />
              </>
            }
          />
          <Route
            path="/ui/accordion"
            element={
              <>
                <PageTitle title="Accordion | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Accordion />
              </>
            }
          />
          <Route
            path="/ui/notifications"
            element={
              <>
                <PageTitle title="Notifications | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Notifications />
              </>
            }
          />
          <Route
            path="/ui/pagination"
            element={
              <>
                <PageTitle title="Pagination | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Pagination />
              </>
            }
          />
          <Route
            path="/ui/progress"
            element={
              <>
                <PageTitle title="Progress | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Progress />
              </>
            }
          />
          <Route
            path="/ui/carousel"
            element={
              <>
                <PageTitle title="Carousel | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Carousel />
              </>
            }
          />
          <Route
            path="/ui/images"
            element={
              <>
                <PageTitle title="Images | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Images />
              </>
            }
          />
          <Route
            path="/ui/videos"
            element={
              <>
                <PageTitle title="Videos | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <Videos />
              </>
            }
          />
          <Route
            path="/auth/signin"
            element={
              <>
                <PageTitle title="Signin | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <SignIn />
              </>
            }
          />
          <Route
            path="/auth/signup"
            element={
              <>
                <PageTitle title="Signup | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <SignUp />
              </>
            }
          />
          <Route
            path="/auth/reset-password"
            element={
              <>
                <PageTitle title="Reset Password | TailAdmin - Tailwind CSS Admin Dashboard Template" />
                <ResetPassword />
              </>
            }
          />
        </Route>
      </Routes>
    </>
  );
}

export default App;
