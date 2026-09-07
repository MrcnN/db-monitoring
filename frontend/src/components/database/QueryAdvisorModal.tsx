import { Fragment } from 'react';
import { Dialog, Transition } from '@headlessui/react';
import { XMarkIcon, SparklesIcon } from '@heroicons/react/24/outline';
import { useQuery } from '@tanstack/react-query';
import { explainQuery } from '../../api/advisor';
import { VisualQueryPlan } from './VisualQueryPlan';

interface QueryAdvisorModalProps {
  isOpen: boolean;
  onClose: () => void;
  databaseId: string;
  query: string;
}

export default function QueryAdvisorModal({ isOpen, onClose, databaseId, query }: QueryAdvisorModalProps) {
  const { data, isLoading, error } = useQuery({
    queryKey: ['advisor', databaseId, query],
    queryFn: () => explainQuery(databaseId, query),
    enabled: isOpen && !!query,
    retry: false,
  });

  return (
    <Transition.Root show={isOpen} as={Fragment}>
      <Dialog as="div" className="relative z-50" onClose={onClose}>
        <Transition.Child
          as={Fragment}
          enter="ease-out duration-300"
          enterFrom="opacity-0"
          enterTo="opacity-100"
          leave="ease-in duration-200"
          leaveFrom="opacity-100"
          leaveTo="opacity-0"
        >
          <div className="fixed inset-0 bg-gray-900 bg-opacity-75 transition-opacity" />
        </Transition.Child>

        <div className="fixed inset-0 z-10 overflow-y-auto">
          <div className="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
            <Transition.Child
              as={Fragment}
              enter="ease-out duration-300"
              enterFrom="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
              enterTo="opacity-100 translate-y-0 sm:scale-100"
              leave="ease-in duration-200"
              leaveFrom="opacity-100 translate-y-0 sm:scale-100"
              leaveTo="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
            >
              <Dialog.Panel className="relative transform overflow-hidden rounded-lg bg-gray-800 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-4xl border border-gray-700">
                <div className="bg-gray-800 px-4 pb-4 pt-5 sm:p-6 sm:pb-4">
                  <div className="sm:flex sm:items-start">
                    <div className="mx-auto flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full bg-indigo-900/50 sm:mx-0 sm:h-10 sm:w-10">
                      <SparklesIcon className="h-6 w-6 text-indigo-400" aria-hidden="true" />
                    </div>
                    <div className="mt-3 text-center sm:ml-4 sm:mt-0 sm:text-left w-full">
                      <Dialog.Title as="h3" className="text-base font-semibold leading-6 text-white flex justify-between items-center">
                        Smart Query Advisor
                        <button onClick={onClose} className="text-gray-400 hover:text-gray-300">
                          <XMarkIcon className="h-6 w-6" />
                        </button>
                      </Dialog.Title>
                      
                      <div className="mt-4">
                        <h4 className="text-sm font-medium text-gray-300 mb-2">Original Query</h4>
                        <div className="bg-gray-900 p-3 rounded text-sm font-mono text-gray-300 overflow-x-auto">
                          {query}
                        </div>
                      </div>

                      {isLoading ? (
                        <div className="mt-6 flex justify-center py-8">
                          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-500"></div>
                        </div>
                      ) : error ? (
                        <div className="mt-6 bg-red-900/20 border border-red-900/50 rounded-md p-4 text-red-400 text-sm">
                          {(error as any)?.response?.data?.error?.message || 'Failed to analyze query. Only SELECT queries are supported.'}
                        </div>
                      ) : data ? (
                        <>
                          <div className="mt-6">
                            <h4 className="text-sm font-medium text-gray-300 mb-2">AI & Heuristics Recommendations</h4>
                            {data.recommendations.map((rec, idx) => (
                              <div key={idx} className="bg-indigo-900/20 border border-indigo-900/50 rounded-md p-3 mb-2 flex items-start text-sm text-indigo-300">
                                <SparklesIcon className="w-5 h-5 mr-2 flex-shrink-0 text-indigo-400" />
                                <span>{rec}</span>
                              </div>
                            ))}
                          </div>
                          <div className="mt-6">
                            <h4 className="text-sm font-medium text-gray-300 mb-2">Raw Execution Plan</h4>
                            <pre className="bg-gray-900 p-3 rounded text-xs font-mono text-gray-400 overflow-x-auto max-h-60 overflow-y-auto mb-4">
                              {JSON.stringify(data.plan, null, 2)}
                            </pre>
                            
                            <VisualQueryPlan planData={data.plan} />
                          </div>
                        </>
                      ) : null}
                    </div>
                  </div>
                </div>
                <div className="bg-gray-900 px-4 py-3 sm:flex sm:flex-row-reverse sm:px-6">
                  <button
                    type="button"
                    className="mt-3 inline-flex w-full justify-center rounded-md bg-gray-800 px-3 py-2 text-sm font-semibold text-gray-300 shadow-sm ring-1 ring-inset ring-gray-600 hover:bg-gray-700 sm:mt-0 sm:w-auto"
                    onClick={onClose}
                  >
                    Close
                  </button>
                </div>
              </Dialog.Panel>
            </Transition.Child>
          </div>
        </div>
      </Dialog>
    </Transition.Root>
  );
}
