import { keepPreviousData, useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Button, ButtonGroup, Text } from "@chakra-ui/react";
import { createColumnHelper, flexRender, getCoreRowModel, useReactTable } from "@tanstack/react-table"
// import axios from 'axios';
import { useState, useCallback } from "react";
import { debounce } from 'lodash';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faTrash } from '@fortawesome/free-solid-svg-icons';
import api from "../components/AxiosInstance";

interface Item {
  serial_number: number,
  rfid_tag: string,
  item_name: string,
  type_ref: string,
  warranty: string,
  sold: boolean,
  modal: number,
  keuntungan: number,
  quantity: number,
  batch: number,
  status: string,
  createdat: Date,
}

const TableScreen = () => {
    const [globalFilter, setGlobalFilter] = useState("")
    const [page, setPage] = useState(1)
    const [search, setSearch] = useState("") 

    const limit: number = 10
    let offset: number = 0

    if (page > 1 ) {
        offset = (page - 1) * limit 
    }

    const queryClient = useQueryClient();

    const debouncedSetSearch = useCallback(
        debounce((query) => setSearch(query), 25),
        []
    );

    const handleSearchChange = (e) => {
        debouncedSetSearch(e.target.value);
    };

    const formatPrice = (price: number) => {
        return new Intl.NumberFormat('id-ID', {
          style: 'currency',
          currency: 'IDR',
          minimumFractionDigits: 0,
        }).format(price);
      };

    const formatDate = (rawDate: Date) => {
        return new Intl.DateTimeFormat('id-ID', {
            day: '2-digit',
            month: '2-digit',
            year: 'numeric',
        }).format(new Date(rawDate));
    };

    const getItems = async (): Promise<{ items: Item[], item_count: number }> => {
        try {
            const response = await api.get<{ items: Item[], item_count: number }>(
                `/api/item/get-items?limit=${limit}&offset=${offset}&search=${search}`, 
                {withCredentials: true});

            return response.data
        } catch (error) {
            console.log(error)
            return { items: [], item_count: 0 }
        }
    }
    
    const { data, error, isLoading, isError } = useQuery<{ items: Item[], item_count: number }>({
        queryKey: ['items', offset, search],    // Refetches when offset/search changes value
        queryFn: getItems,
        placeholderData: keepPreviousData,
    });

    const deleteHandler = async (epc_tag: string) => {
        if (window.confirm("Are you sure you want to delete this item?")) {
            try{
                await deleteItemMutation(epc_tag);
                console.log("Item deleted");
            } catch (error) {
                console.log("Error deleting item:", error)
            }
        }
    }

    const deleteItem = async (epc_tag: string) => {
        try {
            console.log("TAG: ", epc_tag);
            const response = await api.delete<Item>(
                `api/item/delete/${epc_tag}`
            );
            console.log("response:", response);
        } catch (error) {
            console.log("Error deleting:", error);
        }
    };

    const { mutate: deleteItemMutation } = useMutation ({
        mutationFn: deleteItem,
        onSuccess: () => {
            console.log("Item deleted");
            queryClient.invalidateQueries({queryKey: ['items']});
        },
        onError: (error) => {
            console.log("Error deleting item:", error.message);
        }
    });

    const totalRecords: number = data?.item_count ?? 0;
    const canPrevPage: boolean = page > 1;
    const canNextPage: boolean = totalRecords > limit * page;
    let totalPages: number = Math.ceil(totalRecords / limit);
    if (totalPages < 1) {
        totalPages = 1
    }

    const columnHelper = createColumnHelper<Item>()

    const columns = [
        columnHelper.accessor("batch", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    Batch
                </span>
            )
        }),

        columnHelper.accessor("createdat", {
            cell: (info) => (
                <span>{formatDate(info.getValue())}</span>
            ),
            header: () => (
                <span className="flex items-center">
                    Date In
                </span>
            )
        }),

        columnHelper.accessor("serial_number", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    Serial Number
                </span>
            )
        }),

        columnHelper.accessor("rfid_tag", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    RFID Tag
                </span>
            )
        }),

        columnHelper.accessor("type_ref", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    Item Name
                </span>
            )
        }),

        columnHelper.accessor("warranty", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    Warranty
                </span>
            )
        }),

        columnHelper.accessor("status", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    Status
                </span>
            )
        }),

        columnHelper.accessor("modal", {
            cell: (info) => (
                <span>{formatPrice(info.getValue())}</span>
            ),
            header: () => (
                <span className="flex items-center">
                    Modal
                </span>
            )
        }),

        columnHelper.accessor("keuntungan", {
            cell: (info) => (
                <span>{formatPrice(info.getValue())}</span>
            ),
            header: () => (
                <span className="flex items-center">
                    Keuntungan
                </span>
            )
        }),

        columnHelper.accessor("quantity", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    Quantity
                </span>
            )
        }),

        columnHelper.display({
            id: "delete",
            cell: ({ row }) => (
                <button
                    onClick={() => deleteHandler(row.original.rfid_tag)} 
                    className="text-red-600 hover:text-red-800"
                    aria-label="Delete"
                >
                    <FontAwesomeIcon icon={faTrash} />
                </button>
            ),
            header: () => (
                <span className="flex items-center">
                    Delete
                </span>
            )
        }),
    ]

    const items = data?.items || [];

    const table = useReactTable({
        data: items,
        columns,
        state: {
            globalFilter
        },
        getCoreRowModel: getCoreRowModel(),
    })

    if (isLoading) {
        return <div>Loading...</div>;
      }
    
    if (isError) {
        return <div>Error: {(error as Error).message}</div>;
    }

    return (
        <div className="flex flex-col min-h-screen max-w-7xl mx-auto py-2 px-4 sm:px-6 lg:px-5">

            <div className="mb-4 relative">
                <input value={search ?? ""} onChange={handleSearchChange} placeholder="Search..." className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-gray-300 focus:border-gray-300 focus:outline-none"/>
            </div>

            <div className="overflow-x-auto bg-white shadow-md rounded-lg">
                <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                        {
                            table.getHeaderGroups().map((headerGroup) => (
                                <tr key={headerGroup.id}>
                                    {headerGroup.headers.map((header) => (
                                        <th key={header.id} className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            <div>
                                                {flexRender(
                                                    header.column.columnDef.header,
                                                    header.getContext()
                                                )}
                                            </div>
                                        </th>
                                    ))}
                                </tr>
                            ))
                        }
                    </thead>
                        <tbody className="bg-white divide-y divide-gray-200">
                            {table.getRowModel().rows.map((row) => (
                                <tr key={row.id} className="hover:bg-gray-50">
                                    {row.getVisibleCells().map((cell) => (
                                        <td key={cell.id} className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                            {flexRender(cell.column.columnDef.cell, cell.getContext())}
                                        </td>
                                    ))}
                                </tr>
                        ))}
                    </tbody>
                </table>
            </div>

            <Text mb={2} className="mt-2">
                Page {page} of{" "}
                {totalPages}
            </Text>
            <ButtonGroup size="sm" isAttached variant="outline">
                <Button onClick={() => setPage(page - 1)} isDisabled={!canPrevPage} className="mr-1 p-2 rounded-md bg-gray-100 text-gray-600 hover:bg-gray-200 disabled:opacity-50">
                    {"<"}
                </Button>
                <Button onClick={() => setPage(page + 1)} isDisabled={!canNextPage} className="p-2 rounded-md bg-gray-100 text-gray-600 hover:bg-gray-200 disabled:opacity-50">
                    {">"}
                </Button>
            </ButtonGroup>
        </div>
    )
}

export default TableScreen