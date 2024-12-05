import { keepPreviousData, useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import axios from 'axios';
import { useState } from "react";
import { createColumnHelper, flexRender, getCoreRowModel, useReactTable } from "@tanstack/react-table"
import { FaChevronLeft } from 'react-icons/fa';
import { useNavigate } from 'react-router-dom';

interface Type {
    item_type: string,
    price: number,
}

const ItemTypesScreen = () => {
    const [typeName, setTypeName] = useState('')
    const [price, setPrice] = useState(0)

    const queryClient = useQueryClient();
    const navigate = useNavigate()

    const handleTypeNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setTypeName(e.target.value)
        console.log(typeName)
    }

    const handlePriceChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setPrice(parseFloat(e.target.value))
        console.log(price)
    }

    const { data, error, isLoading, isError } = useQuery<{ types: Type[], count: number }>({
        queryKey: ['types'],
        queryFn: async(): Promise<{ types: Type[], count: number }> => {
            const { data } = await axios.get<{ types: Type[], count: number }>(`api/item/get-types`);
            return data
        },
        placeholderData: keepPreviousData,
    });

    const createTypeHandler = async (item_type: string, price: number) => {
        if (item_type && price > 0) {
            if (window.confirm(`Are you sure you want to add this item? \nItem: ${item_type}\nPrice: ${price}`)) {
                const newType: Type = {
                    item_type: item_type,
                    price: price,
                }
                try {
                    await createTypeMutation(newType)
                    console.log("Item created: ", newType)
                    setTypeName('');
                    setPrice(0)
                } catch(error) {
                    console.log("Error creating new type: ", error)
                }
            } else {
                console.log("denied");
            }
        } else {
            return
        }
    }

    const createType = async (item_type: Type) => {
        try {
            console.log("Item Type: ", item_type.item_type);
            console.log("Item Price: ", item_type.price);
            const response = await axios.post<Type>(
                `api/item/register-item-type`,
                item_type
            );
            console.log("response: ", response);
        } catch (error) {
            console.log("error creating type: ", error);
        };
    }

    const { mutate: createTypeMutation } = useMutation({
        mutationFn: createType,
        onSuccess: () => {
            console.log("Type created");
            queryClient.invalidateQueries({queryKey: ['types']});
        },
        onError: (error) => {
            console.log("Error creating new type:", error.message);
        }
    });

    console.log(data)
    const columnHelper = createColumnHelper<Type>()

    const columns = [
        columnHelper.accessor((row, index) => index + 1, {
            id: 'count',  // Give it an ID so we can reference it
            header: () => <span className="flex items-center">#</span>,
            cell: (info) => info.getValue(), // Access the computed value for count
          }),

        columnHelper.accessor("item_type", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className='flex items-center'>
                    Item Type
                </span>
            )
        }), 

        columnHelper.accessor("price", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className='flex items-center'>
                    Price
                </span>
            )
        }),
    ]

    const types = data?.types || []

    const table = useReactTable({
        data: types,
        columns,
        getCoreRowModel: getCoreRowModel(),
    })

    if (isLoading) {
        return <div>Loading...</div>;
      }
    
    if (isError) {
    return <div>Error: {(error as Error).message}</div>;
    }

  return (
    <div className='bg-gray-200'>
        <div className='px-4 pt-2'>
            <button className="px-2 py-1" onClick={() => navigate("/")}>
                <FaChevronLeft size={25} />
            </button>
        </div>

        <div className="flex min-h-screen flex-col justify-top py-20 px-60">
            <div className='text-2xl'> 
                Registered Items List
            </div>

            <div className="overflow-x-auto bg-white shadow-md rounded-md mt-3">
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

            <div className='text-2xl mt-5'> 
                Register New Type
            </div>

            <div className="mt-1.5 relative">
                Name:
                <input value={typeName ?? ""} onChange={handleTypeNameChange} placeholder="Name..." className="w-full pl-5 pr-4 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-gray-300 focus:border-gray-300 focus:outline-none mt-0.5"/>
            </div>

            <div className="mt-3 relative">
                Price:
                <input type="number" value={price ?? ""} onChange={handlePriceChange} placeholder="Price..." className="w-full pl-5 pr-4 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-gray-300 focus:border-gray-300 focus:outline-none mt-0.5"/>
            </div>

            <div>
                <button className="mt-2 px-6 py-2 bg-white border border-gray-300 rounded-md shadow-sm" 
                onClick={() => createTypeHandler(typeName, price)}
                >
                        Add Type
                </button>
            </div>
        </div>
    </div>
  )
}

export default ItemTypesScreen