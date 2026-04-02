import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { Suspense, lazy } from 'react';
import Layout from './components/layout/Layout';
import { LicenseProvider } from './hooks/useLicense';

const Dashboard = lazy(() => import('./pages/Dashboard'));
const ReportDetail = lazy(() => import('./pages/ReportDetail'));
const Compliance = lazy(() => import('./pages/Compliance'));
const Settings = lazy(() => import('./pages/Settings'));
const TestForm = lazy(() => import('./pages/TestForm'));

export default function App() {
  return (
    <LicenseProvider>
      <BrowserRouter>
        <Routes>
          <Route element={<Layout />}>
            <Route index element={<Suspense fallback={<div/>}><Dashboard /></Suspense>} />
            <Route path="/tests/new" element={<Suspense fallback={<div/>}><TestForm /></Suspense>} />
            <Route path="/tests/:name/edit" element={<Suspense fallback={<div/>}><TestForm /></Suspense>} />
            <Route path="/reports/:testName" element={<Suspense fallback={<div/>}><ReportDetail /></Suspense>} />
            <Route path="/compliance" element={<Suspense fallback={<div/>}><Compliance /></Suspense>} />
            <Route path="/settings" element={<Suspense fallback={<div/>}><Settings /></Suspense>} />
          </Route>
        </Routes>
      </BrowserRouter>
    </LicenseProvider>
  );
}
