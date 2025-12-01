import { Dialog, Transition } from '@headlessui/react';
import { Fragment } from 'react';
import { QRCodeSVG } from 'qrcode.react';
import { X, Loader2 } from 'lucide-react';

export default function QRCodeModal({ isOpen, onClose, qrCode, status }) {
    return (
        <Transition appear show={isOpen} as={Fragment}>
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
                    <div className="fixed inset-0 bg-black/80" />
                </Transition.Child>

                <div className="fixed inset-0 overflow-y-auto">
                    <div className="flex min-h-full items-center justify-center p-4 text-center">
                        <Transition.Child
                            as={Fragment}
                            enter="ease-out duration-300"
                            enterFrom="opacity-0 scale-95"
                            enterTo="opacity-100 scale-100"
                            leave="ease-in duration-200"
                            leaveFrom="opacity-100 scale-100"
                            leaveTo="opacity-0 scale-95"
                        >
                            <Dialog.Panel className="w-full max-w-sm transform overflow-hidden rounded-2xl bg-white p-6 text-left align-middle shadow-xl transition-all">
                                <div className="flex justify-between items-center mb-4">
                                    <Dialog.Title as="h3" className="text-lg font-medium leading-6 text-gray-900">
                                        Scan QR Code
                                    </Dialog.Title>
                                    <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
                                        <X className="w-5 h-5" />
                                    </button>
                                </div>

                                <div className="mt-4 flex flex-col items-center justify-center space-y-4">
                                    {qrCode ? (
                                        <div className="p-2 bg-white rounded-lg border-2 border-gray-100">
                                            <QRCodeSVG value={qrCode} size={256} level={"H"} />
                                        </div>
                                    ) : (
                                        <div className="w-64 h-64 flex items-center justify-center bg-gray-50 rounded-lg">
                                            <Loader2 className="w-8 h-8 animate-spin text-blue-500" />
                                        </div>
                                    )}

                                    <p className="text-sm text-gray-500 text-center">
                                        Open WhatsApp on your phone and scan this code to connect.
                                    </p>
                                </div>
                            </Dialog.Panel>
                        </Transition.Child>
                    </div>
                </div>
            </Dialog>
        </Transition>
    );
}
