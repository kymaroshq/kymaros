import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { Suspense, lazy } from 'react';
import Layout from './components/layout/Layout';

const Dashboard = lazy(() => import('./pages/Dashboard'));
const ReportDetail = lazy(() => import('./pages/ReportDetail'));
const Settings = lazy(() => import('./pages/Settings'));

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route index element={<Suspense fallback={<div/>}><Dashboard /></Suspense>} />
          <Route path="/reports/:testName" element={<Suspense fallback={<div/>}><ReportDetail /></Suspense>} />
          <Route path="/settings" element={<Suspense fallback={<div/>}><Settings /></Suspense>} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
